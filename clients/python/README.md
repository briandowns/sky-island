# Sky Island Python Client

The Sky Island Python client provides simple access to the Sky Island functions API.

## Examples

Data ONLY

* Initialize the client
* Run a function to get a Geo Hash

```
client = Client("demo.skyisland.io", 3281)
data = client.function("github.com/mmcloughlin/geohash", "Encode(100.1, 80.9)")

# result: jcc92ytsf8kn
```

All Data

* Initialize the client
* Run a function to get a Geo Hash

```
client = Client("demo.skyisland.io", 3281)
data = client.function("github.com/mmcloughlin/geohash", "Encode(100.1, 80.9)", True)

# result: {u'timestamp': 1513921941, u'data': u'jcc92ytsf8kn'}
```