# Sky Island C Client

A feable attempt at writing a client for Sky Island in C.

## Build

```
cc sky-island.c -lcurl -ljson-c -o sky-func
```

## Examples

```
int main(int argc, char* argv[]) {
  struct client_t* c = malloc(sizeof(struct client_t));
  c->endpoint = "http://demo.skyisland.io:3281/api/v1/function";

  char* url = "github.com/mmcloughlin/geohash";
  char* call = "Encode(100.1, 80.9)";

  function(c, url, call);
  
  free(c);
  return (EXIT_SUCCESS);
}
```