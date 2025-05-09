Instana instrumentation for github.com/IBM/sarama
=====================================================

This module contains instrumentation code for Kafka producers and consumers that use `github.com/IBM/sarama` library starting
from v1.41.0 and above.

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]

Installation
------------

```bash
$ go get github.com/instana/go-sensor/instrumentation/instasarama
```

Usage
-----

For detailed usage examples see [the documentation][godoc] or the following links:
- [Async producer example](./example_async_producer_test.go)
- [Sync producer example](./example_sync_producer_test.go)
- [Consumer_group example](./example_consumer_group_test.go)
- [Consumer_example](./example_consumer_group_test.go)

This instrumentation requires an instance of [`instana.Collector`][Collector] to initialize spans and handle the trace context propagation.
You can create a new instance of Instana collector using [`instana.InitCollector()`][InitCollector].

`instasarama` provides a set of convenience wrappers for constructor functions exported by `github.com/IBM/sarama`. These
wrappers are named the same way as their origins and use the same set of arguments. In most cases it's enough to replace
`sarama` with `instasarama` in the constructor call and append an instance of `instana.TracerLogger` to the argument list.

**Note**: Kafka supports record headers starting from v0.11.0. In order to enable trace context propagation, you need to make sure
that your `(sarama.Config).Version` is set to at least `sarama.V0_11_0_0`.

### Instrumenting `sarama.SyncProducer`

For more detailed example code please consult the [package documentation][godoc] or [example_sync_producer_test.go](./example_sync_producer_test.go).

To create an instrumented instance of `sarama.SyncProducer` from a list of broker addresses use [instasarama.NewSyncProducer()][NewSyncProducer]:

```go
producer := instasarama.NewSyncProducer(brokers, config, collector)
```

[instasarama.NewSyncProducerFromClient()][NewSyncProducerFromClient] does the same, but from an existing `sarama.Client`:

```go
producer := instasarama.NewSyncProducerFromClient(client, collector)
```

The wrapped producer takes care of trace context propagation by creating an exit span and injecting the trace context into each Kafka
message headers. Since `github.com/IBM/sarama` does not use `context.Context`, which is a conventional way of passing the parent
span in Instana Go collector, the caller needs to inject the parent span context using [`instasarama.ProducerMessageWithSpan()`][ProducerMessageWithSpan]
before passing it to the wrapped producer.

### Instrumenting `sarama.AsyncProducer`

Similarly to `sarama.SyncProducer`, `instasarama` provides wrappers for constructor methods of `sarama.AsyncProducer` and expects
the parent span context to be injected into message headers using `instasarama.ProducerMessageWithSpan()`.

For more detailed example code please consult the [package documentation][godoc] or [example_async_producer_test.go](./example_async_producer_test.go).

To create an instrumented instance of `sarama.AsyncProducer` from a list of broker addresses use [instasarama.NewAsyncProducer()][NewAsyncProducer]:

```go
producer := instasarama.NewAsyncProducer(brokers, config, collector)
```

[instasarama.NewAsyncProducerFromClient()][NewAsyncProducerFromClient] does the same, but from an existing `sarama.Client`:

```go
producer := instasarama.NewAsyncProducerFromClient(client, collector)
```

The wrapped producer takes care of trace context propagation by creating an exit span and injecting the trace context into each Kafka
message headers. Since `github.com/IBM/sarama` does not use `context.Context`, which is a conventional way of passing the parent
span in Instana Go collector, the caller needs to inject the parent span context using [`instasarama.ProducerMessageWithSpan()`][ProducerMessageWithSpan]
before passing it to the wrapped producer.

### Instrumenting `sarama.Consumer`

For more detailed example code please consult the [package documentation][godoc] or [example_consumer_test.go](./example_consumer_test.go).

To create an instrumented instance of `sarama.Consumer` from a list of broker addresses use [instasarama.NewConsumer()][NewConsumer]:

```go
consumer := instasarama.NewConsumer(brokers, config, collector)
```

[instasarama.NewConsumerFromClient()][NewConsumerFromClient] does the same, but from an existing `sarama.Client`:

```go
consumer := instasarama.NewConsumerFromClient(client, collector)
```

The wrapped consumer will pick up the existing trace context if found in message headers, start a new entry span and inject its context
into each message. This context can be retrieved with [`instasarama.SpanContextFromConsumerMessage()`][SpanContextFromConsumerMessage]
and used in the message handler to continue the trace.

### Instrumenting `sarama.ConsumerGroup`

For more detailed example code please consult the [package documentation][godoc] or [example_consumer_group_test.go](./example_consumer_group_test.go).

`instasarama` provides [`instasarama.WrapConsumerGroupHandler()`][WrapConsumerGroupHandler] to wrap your `sarama.ConsumerGroupHandler`
into a wrapper that takes care of trace context extraction, creating an entry span and injecting its context into each received `sarama.ConsumerMessage`:

```go
var client sarama.ConsumerGroup

consumer := instasarama.WrapConsumerGroupHandler(&Consumer{}, collector)

// use the wrapped consumer in the Consume() call
for {
	client.Consume(ctx, consumer)
}
```

The wrapped consumer will pick up the existing trace context if found in message headers, start a new entry span and inject its context
into each message. This context can be retrieved with [`instasarama.SpanContextFromConsumerMessage()`][SpanContextFromConsumerMessage] and used
in the message handler to continue the trace.

### Working With Kafka Header Formats

Instana is currently changing how Kafka headers are handled. This change affects how Instana headers are propagated via a producer when a message is sent. 

Starting from [instasarama](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instasarama) v1.24.0, binary headers are no longer used, and you can't set the header format using the environment variable (INSTANA_KAFKA_HEADER_FORMAT). The only available format now is 'string'.

In versions between 1.2.0 and 1.24.0, Instana supports trace correlation headers in both 'binary'(old) and 'string'(new) formats. By default, messages in these versions will include both 'binary' and 'string' headers.

Versions before 1.2.0 will only use 'binary' headers.

See the topic [Kafka header migration](https://www.ibm.com/docs/en/instana-observability/current?topic=references-kafka-header-migration) in Instana's documentation for more information.

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instasarama
[NewSyncProducer]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instasarama?tab=doc#NewSyncProducer
[NewSyncProducerFromClient]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instasarama?tab=doc#NewSyncProducerFromClient
[NewAsyncProducer]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instasarama?tab=doc#NewAsyncProducer
[NewAsyncProducerFromClient]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instasarama?tab=doc#NewAsyncProducerFromClient
[NewConsumer]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instasarama?tab=doc#NewConsumer
[NewConsumerFromClient]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instasarama?tab=doc#NewConsumerFromClient
[WrapConsumerGroupHandler]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instasarama?tab=doc#WrapConsumerGroupHandler
[ProducerMessageWithSpan]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instasarama?tab=doc#ProducerMessageWithSpan
[SpanContextFromConsumerMessage]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instasarama?tab=doc#SpanContextFromConsumerMessage
[Collector]: https://pkg.go.dev/github.com/instana/go-sensor#Collector
[InitCollector]: https://pkg.go.dev/github.com/instana/go-sensor#InitCollector
