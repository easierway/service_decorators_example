# How to gen the thrift

## download & compile boost source
```bash
# Download the the .tar.gz from https://www.boost.org/users/download/#live
# Unpack and go into the directory:
tar -xzf boost_1_50_0.tar.gz
cd boost_1_50_0
# Configure (and build bjam):
./bootstrap.sh --prefix=/some/dir/you/would/like/to/prefix
# Build:
./b2
# Install:
./b2 install
```

## download thrift source
```bash
# Download the the .tar.gz from https://github.com/apache/thrift/tree/0.9.3
# Configure
./configure --with-boost=/dir/for/boost -prefix=/some/dir/you/would/like/to/prefix
# Build
make
# Install
make install

```

## gen source
see: https://thrift.apache.org/tutorial/go
thrift -r --gen go calculator.thrift
