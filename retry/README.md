# retry

custom `github.com/eapache/go-resiliency` retrier, will do the retry on when

* `net.Error`
* Server returns `500,502/504/429` status code
* Service returns `errors.Retry` error
* Service returns `unmarshalling` error
