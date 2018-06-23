//Here is to demostrate how to use "github.com/easierway/service_decorators" to simplify the microservice development
package service_decorators_example

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/easierway/g_met"
	"github.com/easierway/service_decorators"
)

type addServiceRequest struct {
	op1 int
	op2 int
}

func addServiceImpl(req service_decorators.Request) (service_decorators.Response, error) {
	addReq, ok := req.(addServiceRequest)
	if !ok {
		return nil, errors.New("unexpected request format")
	}
	return addReq.op1 + addReq.op2, nil
}

const (
	Host        = "127.0.0.1"
	Port        = "9090"
	NetworkAddr = Host + ":" + Port
)

//calculatorServiceHandler is RPC hanlder
type calculatorServiceHandler struct {
	serviceImpl service_decorators.ServiceFunc
}

//addfallback is the fallback function for CircuitBreakDecorator
func addFallback(req service_decorators.Request, err error) (service_decorators.Response, error) {
	return 0, nil
}

//decorateCoreLogic is to decorate the core logic with the prebuilt decorators
func decorateCoreLogic(innerFn service_decorators.ServiceFunc) (service_decorators.ServiceFunc, error) {
	var (
		rateLimitDec    *service_decorators.RateLimitDecorator
		circuitBreakDec *service_decorators.CircuitBreakDecorator
		metricDec       *service_decorators.MetricDecorator
		err             error
	)
	if rateLimitDec, err = service_decorators.CreateRateLimitDecorator(time.Millisecond*1, 100, 100); err != nil {
		return nil, err
	}

	if circuitBreakDec, err = service_decorators.CreateCircuitBreakDecorator().
		WithTimeout(time.Millisecond * 100).
		WithMaxCurrentRequests(1000).
		WithTimeoutFallbackFunction(addFallback).
		WithBeyondMaxConcurrencyFallbackFunction(addFallback).
		Build(); err != nil {
		return nil, err
	}

	gmet := g_met.CreateGMetInstanceByDefault("g_met_config/gmet_config.xml")
	if metricDec, err = service_decorators.CreateMetricDecorator(gmet).
		NeedsRecordingTimeSpent().Build(); err != nil {
		return nil, err
	}
	decFn := rateLimitDec.Decorate(
		circuitBreakDec.Decorate(
			metricDec.Decorate(
				innerFn)))
	return decFn, nil
}

//decode RPC request
func decodeRPCRequest(req *Request) service_decorators.Request {
	return addServiceRequest{
		op1: int(req.GetOp1()),
		op2: int(req.GetOp2()),
	}
}

//encode the result to RPC response
func encodeRPCResponse(innerResp *service_decorators.Response) int32 {
	return int32((*innerResp).(int))
}

//Add is RPC handler function
func (c *calculatorServiceHandler) Add(ctx context.Context,
	req *Request) (r int32, err error) {
	if err != nil {
		return
	}
	innerResp, err := c.serviceImpl(decodeRPCRequest(req))
	return encodeRPCResponse(&innerResp), err
}

func createCalculatorServiceHandler() (*calculatorServiceHandler, error) {
	decServiceFn, err := decorateCoreLogic(addServiceImpl)
	if err != nil {
		return nil, err
	}
	return &calculatorServiceHandler{decServiceFn}, nil
}

//Start service via RPC endpoint
func startServiceServer(t *testing.T) {
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	serverTransport, err := thrift.NewTServerSocket(NetworkAddr)
	if err != nil {
		t.Error("failed to set tranport", err)
	}
	serviceHandler, err := createCalculatorServiceHandler()
	if err != nil {
		t.Error("failed to create service handler", err)
	}
	processor := NewCalculatorProcessor(serviceHandler)
	server := thrift.NewTSimpleServer4(processor, serverTransport, transportFactory, protocolFactory)
	t.Log("thrift server in", NetworkAddr)
	server.Serve()
}

//Call the service
func startTestClient(t *testing.T) {
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	transport, err := thrift.NewTSocket(net.JoinHostPort(Host, Port))
	if err != nil {
		t.Error("error resolving address:", err)
	}

	useTransport, _ := transportFactory.GetTransport(transport)
	client := NewCalculatorClientFactory(useTransport, protocolFactory)

	if err = transport.Open(); err != nil {
		t.Error("Error opening socket to "+Host+":"+Port, " ", err)
	}
	defer transport.Close()
	ret, err := client.Add(nil, &Request{1, 1})
	t.Logf("Ret=%v, Err=%v\n", ret, err)
}

func Test(t *testing.T) {
	go startServiceServer(t)
	time.Sleep(time.Second * 1)
	go startTestClient(t)
	time.Sleep(time.Second * 1)
}
