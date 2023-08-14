# openapi-extract-schema

go run . ./example/spec.yaml ./out/spec.yaml


## Proposed Naming Rules

### Root schemas

Unique request body: `{verb}{Endpoint}Request` eg 
    - `postV2FooRequest`
Unique response body: `{verb}{Endpoint}{StatusCode}Response` eg:
    - `postV2Foo200Response`
RequestBody shared between more than one verb on same endpoint:
    - `{common}{Endpoint}Request[{n}]` eg:
      - `commonV2FooRequest`
      - `commonV2FooRequest1` (if more than one common request) 
  
ResponseBody shared between more than one verb on same endpoint:
    -  `{common}{Endpoint}{StatusCode}Response[{n}]` eg:
       - `commonV2Foo200Response`
       - `commonV2Foo200Response1` (if more than one common request with this name) 
RequestBody shared between more than one endpoint:
    - {common}Request{n}
Response body shared between more than one endpont:
    - `{common}{StatusCode}Response{n}` eg:
      - common401Response1
Response body shared between more than status code on same endpoint:
    - If has same prefix:
        - commonV2Foo2xxResponse
        - commonV2Foo4xxResponse
    - If different prefixes but none are 2:
      - commonV2FooErrorResponse(n)
Response body shared between more than one status code on different endpoints:
    - as above but remove path


So general form is:

- Request: {verb}{Endpoint}{Request}
- Response: {verb}{Endpoint}{StatusCode}{Response}

If more than one name shares schema:
- prefix with 'common'
- remove verb if different across usages
- remove endpoint if different across usages
- if StatusCode different across usages
  - if 2xx and not 2xx, remove StatusCode
  - If more than one non-2xx replace StatusCode with 'Error'
  - If status codes have identical prefix (eg 200, 201) replace with `2x`
- if more than one of same name exist, suffix with `n`

### Other Schemas
Just use Member name with suffix

Algorithm:

1. Repeatedly:
   1. Look in components/schemas/ for type: object and array/object at first level
   2. For each object found:
      1. call GroupObjects
      2. If it already exists in schemas, add that reference (we may want to backfill our request/response implementation to do this too)3. If not create the new reference 
