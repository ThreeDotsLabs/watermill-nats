# Watermill NATS Pub/Sub
<img align="right" width="200" src="https://threedots.tech/watermill-io/watermill-logo.png">

[![CI Status](https://github.com/ThreeDotsLabs/watermill-nats/actions/workflows/master.yml/badge.svg)](https://github.com/ThreeDotsLabs/watermill-nats/actions/workflows/master.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ThreeDotsLabs/watermill-nats)](https://goreportcard.com/report/github.com/ThreeDotsLabs/watermill-nats)

This is Pub/Sub for the [Watermill](https://watermill.io/) project targeting NATS - primarily streaming use cases via
[JetStream](https://docs.nats.io/nats-concepts/jetstream).

It contains two packages, `/pkg/nats` being the current primary implementation.  This supports JetStream
(default) and also some messaging use cases with core NATS but those aren't a perfect fit for watermill due
to fire and forget nature of the system - acks and nacks for example are implemented as no-ops in that mode.
However if you accept the limitations it does open up some interesting options, for example a system with
publishers that don't require confirmation could use the core NATS mode but send to subjects included in JetStream
subscriptions.  This lies far enough outside the streaming use cases of watermill that its included in an
experimental fashion only but should work if resources are provisioned correctly.  It can be nice to use watermill
as the sole interface to NATS in such a system if eg request/reply is not needed.

There is also an experimental `/pkg/jetstream` package that tracks with the experimental upstream
[JetStream API](https://github.com/nats-io/nats.go/blob/main/jetstream/README.md).  This is targeted for a
stable API within watermill by v2.1 of this package.  Right now it only supports a very minimal configuration
subset but does make an effort to preserve lower level access to the underlying stream provisioning API.  You can
see an example [here](./_examples/jetstream_new.go).

All Pub/Sub implementations can be found at [https://watermill.io/pubsubs/](https://watermill.io/pubsubs/).

Watermill is a Go library for working efficiently with message streams. It is intended
for building event driven applications, enabling event sourcing, RPC over messages,
sagas and basically whatever else comes to your mind. You can use conventional pub/sub
implementations like Kafka or RabbitMQ, but also HTTP or MySQL binlog if that fits your use case.


Documentation: https://watermill.io/

Getting started guide: https://watermill.io/docs/getting-started/

Issues: https://github.com/ThreeDotsLabs/watermill/issues

## Contributing

All contributions are very much welcome. If you'd like to help with Watermill development,
please see [open issues](https://github.com/ThreeDotsLabs/watermill/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+)
and submit your pull request via GitHub.

## Support

If you didn't find the answer to your question in [the documentation](https://watermill.io/), feel free to ask us directly!

Please join us on the `#watermill` channel on the [Gophers slack](https://gophers.slack.com/): You can get an invite [here](https://gophersinvite.herokuapp.com/).

## License

[MIT License](./LICENSE)
