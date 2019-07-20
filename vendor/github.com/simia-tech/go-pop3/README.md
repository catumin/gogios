go-pop3
=======

Golang POP3 library.

Implementation is based on https://github.com/bytbox/go-pop3

The documentation can be found at: http://godoc.org/github.com/simia-tech/go-pop3

The POP3 client can be configured to use a timeout for each command.

To initialize a POP3 client and configure it:

`c, err = pop3.Dial(address, pop3.UseTLS(tlsConfig), pop3.UseTimeout(timeout))`
