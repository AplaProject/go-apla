// Copyright 2016 Josh Baker. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
Package resp provides a reader, writer, and server implementation for the RESP protocol. http://redis.io/topics/protocol

RESP is short for "REdis Serialization Protocol".
While the protocol was designed specifically for Redis, it can be used for other client-server software projects.

RESP has the advantages of being human readable and with performance of a binary protocol.
*/
package resp
