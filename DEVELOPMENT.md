This project contains two packages, `/pkg/nats` being the current primary implementation.  This supports JetStream
(default) and also some messaging use cases with core NATS but those aren't a perfect fit for watermill due
to fire and forget nature of the system - acks and nacks for example are implemented as no-ops in that mode.
However if you accept the limitations it does open up some interesting options, for example a system with
publishers that don't require confirmation could use the core NATS mode but send to subjects included in JetStream
subscriptions.  This lies far enough outside the streaming use cases of watermill that its included in an
experimental fashion only but should work if resources are provisioned correctly.  It can be nice to use watermill
as the sole interface to NATS in such a system if eg request/reply is not needed.

There is also the `/pkg/jetstream` package that tracks with the upstream
[JetStream API](https://github.com/nats-io/nats.go/blob/main/jetstream/README.md). You can
see an example [here](./_examples/jetstream_new.go).
