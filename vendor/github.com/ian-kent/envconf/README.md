envconf
=======

Configure your Go application from the environment.

Supports most basic Go types and works nicely with the built in `flag` package.

```go
package main

import(
  "flag"
  "fmt"
  . "github.com/ian-kent/envconf"
)

func main() {
  count := flag.Int("count", FromEnvP("COUNT", 15).(int), "Count target")
  flag.Parse()
  for i := 1; i <= *count; i++ {
    fmt.Printf("%d\n", i)
  }
}
```

## Licence

Copyright ©‎ 2014, Ian Kent (http://iankent.uk).

Released under MIT license, see [LICENSE](LICENSE.md) for details.
