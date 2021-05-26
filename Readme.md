

##### Add below to your configuration file
```
rateLimitConfig:
  redis:
    host: 0.0.0.0
    port: 6379
  rateLimit:
    - key: /employees
      rate: 5
      unit: minute
    - key: /employees/{id}
      rate: 7
      unit: minute
```

###### Have this line in your Config Go file (for above config)
```
RateLimitConfig rateLimitConfig.RateLimitConfig `yaml:"rateLimitConfig"`
```


###### Use Below code to invoke the validation and return response if error, otherwise continue processing as usual
```
err := sampleRateLimit.ValidateRateLimit(&container.AppConfig.RateLimitConfig , r)
if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte(fmt.Sprintf("Validate Threshold Reached for URL {%s}", r.URL.Path)))
    return
}
```