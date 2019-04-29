## 简介
Nginx 是一个异步框架的Web服务器，也可以用作反向代理，负载均衡和HTTP缓存。

- 官网: <https://www.nginx.org/>
- 文档: <https://nginx.org/en/docs/>
- 源码: <https://nginx.org/download/nginx-1.0.15.tar.gz>
- 开源: <https://github.com/nginx/nginx/tree/branches/stable-1.0>

## 源码安装
```
$ wget http://nginx.org/download/nginx-1.0.15.tar.gz
$ tar xzf nginx-1.0.15.tar.gz && cd nginx-1.0.15
$ ./configure && make
$ sudo make install
$ sudo nginx
$ curl localhost
...
```
## 基本配置
```
http {
  server {
    listen 80;
    server_name localhost;

    location / {
      root html;
      index index.html index.htm;
    }

    location /proxy {
      proxy_pass http://localhost:9094;
    }
  }
}
```
### http 系统配置
### server 服务配置
#### listen 网络连接
#### server_name 服务名称
### location 路由配置

#### root 文件目录
#### index 索引文件
#### proxy_pass 反向代理
