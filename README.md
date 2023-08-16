# openapi-extract-schema

## Overview

A tool for ensuring that all schemas contained in an openapi version 3.x definition are individually specified under the `components.schemas` property.

This ensures that code generation tools such as https://github.com/deepmap/oapi-codegen can generate all entities with named structs.

Usage:

`go run . <input-path> <output-path>`

## Operation

The tool does the following:
1. Searches `paths.{endpoint}.{verb}.requestBody.content.{content-type}.schema` and moves inline definitions to `components.schemas`
1. Searches `paths.{endpoint}.{verb}.responses.{statusCode}.content.{content-type}.schema` and moves inline definitions to `components.schemas`
2. Repeatedly (until no more found):
   1.  searches `components.schemas.{name}.properties.{name}.[?(@type=='object')]` and moves inline definitions to schemas
   1.  searches `components.schemas.{name}.properties.{name}.items[?(@type=='object')]` and moves inline definitions to schemas

Where schemas are identical, a single symbol and definition is used.

## Naming Rules

See `./internal/path_test.go` but in summary:

1. For a Request body, if the schema is:
   1. unique: `{Verb}{Path}Request`
   2. used for one path but multiple verbs: `Common{Path}Request` 
   3. used for one verb but multiple paths `Common{Verb}Request`
2. For a Response body, if the schema is:
   1. unique: `{Verb}{Path}{StatusCode}Response`
   2. used for one path and status code but multiple verbs: `Common{Path}{StatusCode}Request` 
   3. used for one verb and status code but multiple verbs `Common{Verb}{StatusCode}Request`
   4. if status codes have same prefix `{StatusCode}` is given as `{prefix}xx`
   5. if status codes with different prefixes, `{StatusCode}` is omitted
3. For an embedded schema if the schema is:
   1. unique: `{ContainingObject}{PropertyName}`
   2. duplicated `Common{PropertyNameOfFirstUse}`
3. For an embedded array schema if the schema is:
   1. unique: `{ContainingObject}{PropertyName}Item`
   2. duplicated `Common{PropertyNameOfFirstUse}Item`

In any of the above cases, if the chosen name already exists, an index suffix is added.
