# go-blog-rate-limiting (github.com/antonio-alexander/go-blog-rate-limiting)

The goal of this repo is to attempt to implement a handful of rate limiting solutions, demo different algorithms and try to create and validate some use cases.

## Helpful Links

- [https://blog.logrocket.com/rate-limiting-go-application/](https://blog.logrocket.com/rate-limiting-go-application/)
- [https://gobyexample.com/rate-limiting](https://gobyexample.com/rate-limiting)
- [https://betterprogramming.pub/4-rate-limit-algorithms-every-developer-should-know-7472cb482f48](https://betterprogramming.pub/4-rate-limit-algorithms-every-developer-should-know-7472cb482f48)
- [https://docs.nginx.com/nginx/admin-guide/load-balancer/http-load-balancer/](https://docs.nginx.com/nginx/admin-guide/load-balancer/http-load-balancer/)
- [https://github.com/uber-go/ratelimit](https://github.com/uber-go/ratelimit)
- [https://pkg.go.dev/golang.org/x/time/rate#Limiter](https://pkg.go.dev/golang.org/x/time/rate#Limiter)
- [https://stackoverflow.com/questions/1321878/how-to-prevent-favicon-ico-requests](https://stackoverflow.com/questions/1321878/how-to-prevent-favicon-ico-requests)
- [https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/429](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/429)
- [https://pspdfkit.com/blog/2018/how-to-use-docker-compose-to-run-multiple-instances-of-a-service-in-development/](https://pspdfkit.com/blog/2018/how-to-use-docker-compose-to-run-multiple-instances-of-a-service-in-development/)
- [https://gist.github.com/xameeramir/a5cb675fb6a6a64098365e89a239541d](https://gist.github.com/xameeramir/a5cb675fb6a6a64098365e89a239541d)

## Use Cases

These are a collection of use cases I want to test and confirm:

- As an application, if I attempt to access an endpoint that's rate limited at 1Hz at 2Hz, I'll receive my responses at 1Hz
  - Constraints are that the messages will run for a known maximum time (to avoid having to use infinite buffers and wait an indeterminate amount of time)

- As an application, if I attempt to access an endpoint that's rate limited at 1Hz at 2Hz for half a second, I'll receive my responses at 1Hz

- As an application, if I attempt to access an endpoint that's rate limited at 1Hz and send a burst of requests, I'll receive my responses at 1Hz
  - Constraints are that the burst has a certain size upon which we'll do something if the burst size is exceeded

- As an application, If I attempt to access an endpoint that's rate limited and I send multiple requests with a payload that will take a significant amount of time, I'll be rate limited instead
  - This assumes some way to weight a request dependent on the payload contents

- As a rate limited application, if multiple applications attempt to access an endpoint that's rate limited, but don't exceed their specific rates, how can I ensure that in aggregate I don't exceed the rate limit?
  - This attempts to capture a situation like our bandwidth being 10Hz and we have 11 applications sending data at 1 Hz

## Algorithm(s)

In this section, we'll discuss a number of algorithms that can be used for rate limiting.

Each of the algorithms below attempt to solve a use case with an algorithm; I think that they serve as a good base to understand what you can do, but I think it's important to consider the following questions when trying to build your own algorithm or determine what algorithm you want to use:

1. Are you using rate limiting to shape traffic (and by extension cpu/memory) or are you using it to enforce how you want your API used?
2. How can you quantify your max resource usage from the perspective of your API?
3. Do you need application specific information to properly implement your algorithm?
4. Is your environment also performing load balancing that could affect your algorithm?

### Token Bucket

This algorithm is straight forward, it has the following "rules":

- each identifiable entity has a bucket assigned to them with a set number of tokens
- each time a request is processed, a token is removed from the bucket
- the tokens in the bucket are replenished at a known interval
- if the bucket is empty when a new request is received, the request is discarded

This solution doesn't quite solve the rate limiting problem; although it'll limit the total number of requests that can be done in a given time interval, it doesn't enforce a given rate; it's still possible to make a lot of requets in a short period of time which can cause CPU/Memory spikes. In addition if a request takes longer than the interval to replenish tokens, they can have a lot of requests in flight.

> In short, the simplicity of this algorithm belies the simplicity of circumventing it. this algorithm can be slightly modified such that you replenish a token per interval or something along those lines, but different forms of the same problem mentioned above remain.

Also keep in mind that the algorithm is only impelmented on the front end and doesn't care what happens to a request afterwards; so there's room to have a lot of requests in flight still.

### Leaky Bucket

This algorithm attempts to completely solve the problem inherent in the token bucket: that you can't maintain a given rate when you only control the number of requests that can be sent. In this solution, these are the rules:

- each identifiable entity has a bucket with a maximum number of requests that can be processed
- if the bucket is full, new requests are discarded
- the requests in the bucket are processed at a set interval

This definitely mitigates the primary failure of the Token Bucket algorithm, but severely impacts the response time of your service _artificially_, if someone sends multiple requests at a rate faster than the process interval, they'll always get their responses back artificially slower. In addition, depending on the payload, because this solution has to cache reqeusts, not only do we have coupling between requests (because they share a queue); we also have to store more data in memory at a given time.

> This solution specifically solves the problem of smoothing or normalizaing your cpu/memory usage by processing requests at a known rate. Unfortunately, it's still very blind in that it assumes that all requests are the same and doesn't necessarily account for concurrent requests.

### Fixed Window

The fixed window algorithm splits time into a window of a given duration and only allows a certain number of requests within that window. The rules are as follows:

- each identifiable entity can execute a set number of requests within a given window
- each request in a given window will reduce the "bucket" by one
- if the maximum number of requests have been made in a given window, the requests are discarded and they'll have to wait until the next window

Ignoring the time math around identifying if you're in an old window or a new window; this algorithm too is relatively easy. It has a similar problem to the token bucket algorithm in that you can still be pretty bursty around the time when a window end and another window begins.

> This algorithm won't do anything to mitigate cpu spikes around the interval window

### Sliding Window

A sliding window algorithm is more complicated that previous solutions, but is like an amalgamation of all of them put together specifically to solve the use case of enforcing a given number of requests per unit of time. It accomplishes this by having enough history to determine how many requests were received a minute in the past.

## Implementation

Below i'll show how to implement rate limiting from the perspective of the server and from the perspective of the client.

### Server

The server implementation is relatively straight forward if you ignore the algorithms. In general, when the endpoint is executed, you'll execute the rate limiting algorithm and from there it can provide some feedback as to what to do; the general options will be:

- discard the request and provide a 429 TOO MANY REQUESTS status code, populate Retry-After if the rate limiting is based on time
- cache the request and process it at a slower rate (using some kind of queue)
- process the request at the same time it was sent

> Although it's totally possible to cache requests and process them at a slower rate that they came in at, it can get terrible complex and if done correctly, you'd still have an upper limit where you'd need to start discarding requests as they too can affect the problem you're trying to mitigate with rate limiting

### Client

The client is relatively straight forward, we can use the Request/Response contracts to affect how the requests are processes and the use of context.WithTimeout() allows us to cancel requests that run long. The client itself has two modes, "single_request" and "multiple_requests" that can be used to affect how many requests are sent.

The single_request mode is as advertised, it will send a single request with a given id, application id, weight and wait. The server will attempt to process it and send back a response. The multiple_requests mode can be used to send a number of simultaneous requests at a configured rate; you can configure the number of requests per interval, the number of applications as well as the weight of each request.

In both modes, there is also the ability to configure retry logic. Within this logic if a 429 too many requests is received, it'll look for the Retry-After header and use that (in milliseconds) to attempt to retry up to the configured maximum number of retries.

## Proof of Concept

Application that has a single endpoint; this endpoint is rate limited and the application can be configured at runtime to choose which algorithm is used and populate the configuration within. In addition, it should be possible to create two instances of this application and using nginx? load balance across these two instances but maintain a common rate limit.

Things I want to try to confirm:

- increased cpu usage with concurrent interactions
- confirmation that load balancing across two contains can reduce cpu usage (without rate limiting)
- confirmation that rate limiting can reduce overall cpu usage (can do this with a couple scenarios like bursts at a given rate where every second you send 100 requests)
- confirmation of the ability to "cancel" using contexts requests that have been rate limited, but have timedout.

I think for the functionality, the endpoint we rate limit is simply a wait function; you provide it with a payload of id and wait (time.Duration) and when you hit the endpoint, it'll wait that given time and then provide you with a response of id and wait time. This should use context so it can immediately return if the request is cancelled. You can send this with a time of 0 for it to return immediately.

## Architecture Concerns

I think the biggest misconception I had attempting to put together this repository was that I could look at rate limiting in a vacuum and apply it practically. I don't think that's possible. Rate limiting, in short, is an exercize in understanding what resources you have available and enforcing a limit such that if all the resources that are available are used, your system doesn't try to use more resources (potentially causing the system to fail closed).

In control systems, there's a tool called a proportional–integral–derivative (PID) controller. This logic will take a setpoint, a control variable and configuration and will change the control variable until the setpoint is reached. PIDs are super useful for things like temperature control or cruise control within a car. PIDs can be updated to include multiple control variables to provide even more functionality. Rate limiting, from the perspective of scaling (think kubernetes), introduces two control variables which can fight each other.

Rate limiting will shape your CPU and memory usage, when fully loaded, it should mean that your memory and CPU usage remain static on average while fully loaded. Kubernetes often has triggers to horizontally scale when your memory or cpu usage reaches a certain threshold and scale down when it drops below a given threshold. Adding rate limiting could cause you NOT to horizontally scale and instead drop requests when you actually have the resources.

As a result, it's not so simple to "just" apply rate limiting, you have to find someway to quantify your resources and tune it such that rate limiting will do its job, but not reduce the effectiveness of kubernetes.

I think it's important to indicate that rate limiting has __TWO__ jobs: (1) to limit the number of requests being handled over a period of time and (2) to enforce how you want applications to use your API. Generally, when you use any API, you have an API key or some unique identifier that allows auditing of any given request: "this request was done by person X".

## Alternative Solutions for Application Based Rate Limiting

In this section, I want to provide some alternate solutions for rate limiting that contrast with "rolling your own". I think it's reasonable to roll your own solution sometimes, but rate limiting is relatively simple (depending on your use case) and there are a lot of solutions already.

It's possible to use NGINX for rate limiting (and load balancing). NGINX is an open sourced solution common for proxy, load balancing and rate limiting. Although there is a limit to what the algorithm is aware of, this can easily serve as your minimum viable product (mvp).

It's also possible to load balance using the sacle directive and docker compose (look at the Makefile)

## Frequently Asked Questions (FAQ)

Should rate limiting only affect the initial endpoint or should it also be taken into consideration for downstream requests?

> This is a poorly constructed question/scenario, but the short version is no, rate limiting should only take into account the endpoint that was accessed. The longer version is that it's possible, but probably not what you want. There's the difficulty of tracking a request through all its downstream endpoints to shape the traffic AND whether or not it matters. IF you call one endpoint and it calls 10 other endpoints, you shouldn't be throttled by those other 10 calls you can't control.

Should you roll your own solution?

> No, you should avoid rolling your own solution by any means possible unless: (A) you have an application specific need that you can't build into an off the shelf solution, (B) want to provide an API to control rate limiting on the fly or (C) have an algorithm that's not available. I think creating the algorithm and logic is relatively simple, it's just that it's easy to get wrong if you don't have extensive tests

How does rate limiting interact with cancelled requests?

> I think that any API that attempts to implement rate limiting HAS to be able to respond to cancelled requests. Most rate limiting algorithms fall apart when it comes to bursty data, specifically bursty data around the time where the algorithm reaches its criteria to stop limiting requests. IF you're not taking into account cancelled requests, it's possible to appear as if you have a lot of requests in flight, but in actuality, the consumer is no longer waiting for a response (and your application is wasting cpu/memory). While not every possible process can dynamically respond to a cancelled request, not doing so if you can is leaving money on the table.

![archer_flex_account](./_images/archer_flex_account.gif)

How is a given algorithm affected by multiple instances of a given application?

> If your application can scale and you've implemented rate limiting within the thing that scales rather than the thing on the outside (e.g. NGINX); then you have to have a common data store such as redis or a database such that each instance is operating with the same information such that you experience the _same_ rate limiting regardless of what instance you're connected to.
