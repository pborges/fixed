# Fixed width marshal/unmarshal
## not production ready

- [x] string
- [x] int, int32, int64
- [x] time.Time
- [x] custom (MarshalFixed,UnmarshalFixed)
- [ ] pointers
    - [x] limited pointer support on strings and time.Time objects but I need to refactor this to be able to support all pointer types more easily
- [ ] maps
- [ ] array
- [ ] nested