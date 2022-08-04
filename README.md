# user-sso

基于user服务的sso sdk，降低接入成本，因为暂时没有精力实现oidc，暂时用最简单的方式实现。
但是基本流程差不多，本质是把用户信息作为resource，使用access token获取用户信息。


- new client,指定oauth2要求的参数和user endpoint
- new token conf，可以encode和decode，支持gzip压缩以减小体积
