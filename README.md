## Proxy сервер для HTTP/HTTPS запросов

Выполнены 1 и 2 пункты задания

#### Build and Run
    
    docker build . -t proxy
    docker run --publish 8080:8080 --name proxyserver -t proxy