# chat-filter-grpc-plugin-server-go

```mermaid
flowchart LR
   subgraph AccelByte Gaming Services
   KF[Kafka]
   subgraph Event Listener Deployment
   KB[Kafka Bridge]   
   SV["gRPC Server\n(YOU ARE HERE)"]   
   end   
   KB --- SV
   KF --- KB
   end   
   
```

`AccelByte Gaming Services` events can be extended by using custom functions implemented in a `gRPC server`. 
If configured, custom functions in the `gRPC server` will be called by `Kafka Bridge` to run specific sequence 
specifically required by a game.

The `Kafka Brdige` and the `gRPC server` can actually communicate directly. 
However, additional services are necessary to provide **security**, **reliability**, **scalability**, and **observability**. 
We call these services as `dependency services`. 
The [grpc-plugin-dependencies](https://github.com/AccelByte/grpc-plugin-dependencies) repository is provided 
as an example of what these `dependency services` may look like. It contains a docker compose which consists of 
these `dependency services`.

> :warning: **grpc-plugin-dependencies is provided as example for local development purpose only:** 
> The dependency services in the actual gRPC server deployment may not be exactly the same.

## Overview

This repository contains a `sample event listener app` written in `Go`. 
It provides a simple custom app that will listen to `userLoggedIn` event, and then tried to grant an entitlement to 
the said user. Preconfigured store and `IAM Client ID` is needed to run this app.

This sample app also shows how this `gRPC server` can be instrumented for better observability.
It is configured by default to send metrics, traces, and logs to the observability `dependency services` 
in [grpc-plugin-dependencies](https://github.com/AccelByte/grpc-plugin-dependencies).

## Project Structure

```text
...
├── main.go
└── pkg
    ├── pb                      # Where the generated protobuf located
    ├── proto                   # Where we put the protobuf for AGS event specification
    └── server
        ...
        ├── loginListener.go     # Where we put custom logic that will get called when the event we interested got invoked
        ...
...
```

As for AGS event specification, it can be obtained from [here](https://github.com/AccelByte/accelbyte-api-proto/tree/main/asyncapi/accelbyte)
in this case we're only interested on "user logged in event", so we only put the event specification for IAM in
this sample app


## Prerequisites

Before starting, you will need the following.

1. Windows 10 WSL2 or Linux Ubuntu 20.04 with the following installed.

   a. bash

   b. curl

   c. docker v23.x
   
   d. docker-compose v2.x
   
   e. docker loki driver
      ```  
      docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions
      ```
   
   f. make
   
   g. go v1.20
   
   h. jq
   
   i. [postman](https://www.postman.com/)

   j. [grpcurl](https://github.com/fullstorydev/grpcurl)


2. A local copy of [grpc-plugin-dependencies](https://github.com/AccelByte/grpc-plugin-dependencies) repository.

   ```
   git clone https://github.com/AccelByte/grpc-plugin-dependencies.git
   ```

3. Access to `AccelByte Gaming Services` demo environment.

   a. Base URL: https://demo.accelbyte.io.
   
   b. [Create a Game Namespace](https://docs.accelbyte.io/esg/uam/namespaces.html#tutorials) 
      if you don't have one yet. Keep the `Namespace ID`.
   
   c. [Create an OAuth Client](https://docs.accelbyte.io/guides/access/iam-client.html) with `confidential` client type. 
      Keep the `Client ID` and `Client Secret`.
   
      - ADMIN:NAMESPACE:*:USER:*:ENTITLEMENT [UPDATE]

4. [Extend Helper CLI](https://github.com/AccelByte/extend-helper-cli) to upload the app to AGS
   
   - Note that to use the tool you'll need an AGS account, be sure to follow the docs on the github link above

5. Published AGS Store

   - Note the item id, as this app will grant that item as an entitlement after a user 
     in a certain namespace successfully logged in

## Setup

To be able to run this sample app, you will need to follow these setup steps.

1. Create a docker compose `.env` file by copying the content of [.env.template](.env.template) file.
2. Fill in the required environment variables in `.env` file as shown below.

   ```
   AB_BASE_URL=https://demo.accelbyte.io      # Base URL of AccelByte Gaming Services demo environment
   AB_CLIENT_ID='xxxxxxxxxx'                  # Use Client ID from the Prerequisites section
   AB_CLIENT_SECRET='xxxxxxxxxx'              # Use Client Secret from the Prerequisites section
   AB_NAMESPACE='xxxxxxxxxx'                  # Use Namespace ID from the Prerequisites section
   PLUGIN_GRPC_SERVER_AUTH_ENABLED=false      # Enable or disable access token and permission verification
   ITEM_ID_TO_GRANT='xxxxxxxxxxx'             # Item id from a published store we noted previously
   ```

   > :warning: **Keep PLUGIN_GRPC_SERVER_AUTH_ENABLED=false for now**: It is currently not
   supported by `AccelByte Gaming Services`, but it will be enabled later on to improve security. 

## Building

To build this sample app, use the following command.

```
make build
```

## Running

To (build and) run this sample app in a container, use the following command.

```
docker-compose up --build
```

## Testing

### Functional Test in Local Development Environment

The custom function in this sample app can be tested locally using [grpcui](https://github.com/fullstorydev/grpcui).

1. Run the `dependency services` by following the `README.md` 
   in the [grpc-plugin-dependencies](https://github.com/AccelByte/grpc-plugin-dependencies) repository.

   > :warning: **Make sure to run dependency services with mTLS disabled for now**: 
   > It is currently not supported by `AccelByte Gaming Services`, but it will be enabled later on to improve security. 

2. Run this `extend-event-listener-go` sample app.

3. Install `grpcui`, please refer to the official doc on the installation, and then run this command

   ```shell
   grpcui -plaintext localhost:6565
   ```

   with `localhost:6565` is the address for our `extend-helper-listener-go`, now go to the Web UI with 
   the URL generated by `grpcui`

4. Now in `grpcui` send kafka event's sample you're interested in, in this case we're interested on `userLoggedIn` event,
   So, we're using this sample payload in [here](https://docs-preview.accelbyte.io/gaming-services/api-events/iam-account/#message-userloggedin)


   ```json
   {
     "payload": {
        "userAccount": {
           "userId": "string",
           "emailAddress": "string",
           "country": "string",
           "namespace": "string"
        },
        "userAuthentication": {
           "platformId": "string",
           "refresh": true
        }
     },
     "id": "string",
     "version": 0,
     "name": "string",
     "namespace": "string",
     "parentNamespace": "string",
     "timestamp": "2019-08-24T14:15:22Z",
     "clientId": "string",
     "userId": "string",
     "traceId": "string",
     "sessionId": "string"
   }
   ```

   You can change the field value you're interested to suits your need, e.g. `namespace` , `userId`, etc

   Then, you can use `grpcui` this way, and please ensure you're selecting the service name and method name, 
   you're interested in. Then hit `Invoke` to send the request.

   ![grpcui request](./docs/grpcui-request.png)


   > Note. if you're interested on other events' sample json, you can find it here
   > https://docs-preview.accelbyte.io/gaming-services/api-events/

 
5. If successful, you will see in the response as follows, and also can see the granted entitlement
   to the user you're tested.
   
   ![grpcui response](./docs/grpcui-response.png) 

   ![granted entitlement](./docs/granted-entitlement.png)

### Integration Test with AccelByte Gaming Services

After passing functional test in local development environment, you may want to perform
integration test with `AccelByte Gaming Services`. Here, we are going to deploy our sample app to AGS.

1. Download and setup [extend-helper-cli](https://github.com/AccelByte/extend-helper-cli/)

2. Create event listener app, please refer to the docs in [here](https://docs-preview.accelbyte.io/gaming-services/services/customization/)

3. Do a docker login using `extend-helper-cli`, please refer to its documentation

4. Upload the image

```
make imagex_push IMAGE_TAG=v0.0.1 REPO_URL=xxxxxxxxxx.dkr.ecr.us-west-2.amazonaws.com/accelbyte/justice/development/extend/xxxxxxxxxx/xxxxxxxxxx 
```

> Note. the REPO_URL is obtained from step 2 in the app detail on the 'Repository Url' field