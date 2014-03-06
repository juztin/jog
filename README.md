Jog
===

Jog is a simple log implementation that outputs JSON data. Loggers may be implemented to send log data to standard HTTP, OAUTH, etc.  

The standard log package can be used with jog, by calling log.SetOuput with a jog writer. 

Installation
--------------

```
$ go get bitbucket.org/juztin/jog
```

Usage
-----

    // Log just like normal
    log.Println("domo arigato mr roboto")
    /* This will produce the following:
     * {
     *   "message": "domo arigato mr robot",
     *   "level": "info",
     *   "file": "/home/you/thisfile.go",
     *   "line": 42,
     *   "time": "2014-03-06T19:38:32.834223448Z"
     * }
     */
    
    // Log with a level (defaults to INFO)
    log.Println(Message{jog.ERROR, err})
    
    // Log with custom JSON
    log.Println(CustomMessage{Name: "Jack", Age: 42, Message: "Failed Auth"})
    /* {
     *   "data": {
     *       "name": "Jack",
     *       "age": 42,
     *       "message": "Failed Auth"
     *   },
     *   "level": "info",
     *   "file": "/home/you/thisfile.go",
     *   "line": 42,
     *   "time": "2014-03-06T19:38:32.834223448Z"
     * }
     */

    // Log with custom JSON and Level
    log.Println(CustomMessage{jog.WARNING, "Jack", 42, "Failed Auth"})
    /* {
     *   "data": {
     *       "name": "Jack",
     *       "age": 42,
     *       "message": "Failed Auth"
     *   },
     *   "level": "warning",
     *   "file": "/home/you/thisfile.go",
     *   "line": 42,
     *   "time": "2014-03-06T19:38:32.834223448Z"
     * }
     */
     

With the above `CustomMessage` type the `fmt.Stringer` interface needs to be implemented to just return the JSON for the object.  

*(This is due to the log package functions performing `fmt.Sprintf(format, v...)` to PrintX calls)*

    func (m CustomMessage) String() string {
        b, _ := json.Marshal(m)
        return string(b)
    }


Customizing
=========
To implement your own custom Logger *(to use OAUTH2, Loggly, etc.)* take a look at at the Basic logger implementation within the `loggers` sub-package

License
----

Simplified BSD
