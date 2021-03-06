package SampleRateLimit

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/skshukla/sampleRateLimit/config"
	appRedis "github.com/skshukla/sampleRateLimit/redis"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func validateConfigurationsArePresent(appConfig *config.RateLimitConfig)  error{
	if appConfig.Redis.Host == "" {
		log.Fatal("Redis Host config not present")
	}
	if appConfig.RateLimit == nil {
		log.Fatal("Rate Limit property not present")
	}
	return nil
}

func ValidateRateLimit(appConfig *config.RateLimitConfig, r *http.Request) error{
	validateConfigurationsArePresent(appConfig)
	var redisConn redis.Conn = appRedis.GetRedisConn()
	defer redisConn.Close()
	threshold, unit := getRateLimitThreshold(appConfig, r.URL.Path)
	if threshold == -1 {
		return nil
	}
	ip := getIpFromRequest(r)
	bucketNameForRateLimiting := getDynamicBucketNameForRateLimiting(ip, r.URL.Path, unit)
	val, err := redis.Int(redisConn.Do("GET", bucketNameForRateLimiting))
	var newVal int
	if err != nil {
		newVal = 1
		redisConn.Do("SET", bucketNameForRateLimiting, newVal)
		redisConn.Do("EXPIRE", bucketNameForRateLimiting, 5 * 60)
	} else {
		if val >= threshold {
			err := errors.New(fmt.Sprintf("Max Rate Limit threshold {%d} Reached from Bucket {%s}, Please try after some time", threshold, bucketNameForRateLimiting))
			return err
		}
		newVal = val + 1
		redisConn.Do("SET", bucketNameForRateLimiting, newVal)
	}
	fmt.Print(fmt.Sprintf("Stored the value {%d} into Redis for bucket {%s}\n", newVal, bucketNameForRateLimiting))
	return nil
}

func getDynamicBucketNameForRateLimiting(IP string, path string, unit string)  string {
	var x int64  = 60
	if unit == "second" {
		x = 1
	}
	return IP + "_" + path + "_" + strconv.FormatInt(time.Now().Unix() / x, 10)
}

func getRateLimitThreshold(appConfig *config.RateLimitConfig, url string) (int, string) {
	for _, x := range appConfig.RateLimit {
		if x.Key == url {
			return x.Rate, x.Unit
		}
	}
	return -1, "minute"
}

func getIpFromRequest(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}

	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}

	return strings.Split(IPAddress, ":")[0]
}