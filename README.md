# Http Mock Server

A simple HTTP mock develpment server.

- Auto reloads when the configuration or a file in the root directory changed

### Config
By default the mock server reads from routes.yaml
This file should contain all the routes

> !Important all the routes are checked from top to bottom as defined in the routes.
> When a match is found the server returns that response and stops the matching

#### > path 
The request uri path, this is a wild card regex matched path.

#### > method
Http method
- GET 
- POST 
- PUT 
- DELETE

#### > query
The query params after the path like `?foo=bar&baz=foo`
This is a hashmap where the value of the param will be matched against the.

#### > headers
Matcher for the request headers.
This is a hashmap where the value of the header will be matched against the regex pattern.

#### > body
This is the raw body of the posted content of a POST, PUT or PATCH request.
The contents will be matched against the regex pattern.

### Response

#### > code
The HTTP response code e.g. 200 / 302 / 500

#### > body
This is the returned response.
This could be a filename, if the filename exists the contents of the file will be returned.
If a file is loaded the content-type will be determined by the filename.
When no file is found the contents of the body will be returned instead

#### > headers
The response headers returned by the request

### Example /data/routes.yaml
```yaml
routes:
  - path: /test/path
    method: GET
    response:
      body: ./test.json
      code: 200

  - path: /test/regex/.*
    method: GET
    query:
      foo: bar
    response:
      body: >
        This is plain text
      code: 200
      headers:
        "Content-Type": "application/text"

  - path: /test
    method: POST
    headers:
      "Content-Type": ".*/x-(w+)-form-urlencoded"
    body: >
      field1=foo&field2=test
    response:
      body: ./response.json
      code: 201
      headers:
        "Location": "http://www.test.com"
``` 

### Docker run a local instance
See [https://hub.docker.com/r/mbict/http_mock_server](https://hub.docker.com/r/mbict/http_mock_server) for more info.

Build
```bash
docker build --target app -t mbict/http_mock_server .
```

Run
```bash
docker run --rm -v $(PWD)/data:/data  -p 8080:8080 mbict/http_mock_server 
```


