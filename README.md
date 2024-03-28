## TODO
- [ ] move over the chat library code from go-discoverybot
- [ ] move over the feed scouring code from psyduck-etl/ifunny
- [ ] basic crawling stuff: getting content ( from feeds, explore, timelines ), observe users, comments
- [ ] channels for streaming crawl-able content ( think iterators from the node library ), using golang-set `Iterator`s might make sense here

## Not TODO 
- auth - supply your own basic / bearer token 
- writable api - might accept patches, but this intends to be primarily a read-only package 
- extra features - this library only aims to support core iFunny features