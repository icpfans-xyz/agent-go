package polling

import (
	"fmt"
	"time"

	"github.com/icpfans-xyz/agent-go/agent"
	"github.com/icpfans-xyz/agent-go/agent/http"
	"github.com/icpfans-xyz/agent-go/principal"
)

type Predicate = func(*principal.Principal, *agent.RequestId, http.RequestStatusResponseStatus) bool

const FIVE_MINUTES = 5 * 60 * time.Second

func DefaultStrategy() PollStrategy {
	return chain(conditionalDelay(once(), 1000), backoff(1000, 1.2), timeout(FIVE_MINUTES))
}

/**
 * Predicate that returns true once.
 */
func once() Predicate {
	first := true
	return func(p *principal.Principal, ri *agent.RequestId, rsrs http.RequestStatusResponseStatus) bool {
		if first {
			first = false
			return true
		}
		return false
	}
}

/**
 * Delay the polling once.
 * @param condition A predicate that indicates when to delay.
 * @param duration The amount of time to delay.
 */
func conditionalDelay(condition Predicate, duration time.Duration) PollStrategy {
	return func(p *principal.Principal, ri *agent.RequestId, rsrs http.RequestStatusResponseStatus) error {
		if condition(p, ri, rsrs) {
			time.Sleep(duration)
		}
		return nil
	}
}

/**
 * Reject a call after a certain amount of time.
 * @param duration Time before the polling should be rejected.
 */
func timeout(duration time.Duration) PollStrategy {
	end := time.Now().Add(duration)
	return func(p *principal.Principal, ri *agent.RequestId, rsrs http.RequestStatusResponseStatus) error {
		if time.Now().After(end) {
			return fmt.Errorf("Request timed out after %d; Request ID:%v; status:%s", duration, *ri, rsrs)
		}
		return nil
	}
}

/**
* A strategy that throttle, but using an exponential backoff strategy.
* @param startingThrottle The throttle to start with.
* @param backoffFactor The factor to multiple the throttle time between every poll. For
*   example if using 2, the throttle will double between every run.
 */
func backoff(startingThrottle time.Duration, backoffFactor float32) PollStrategy {
	currentThrottling := startingThrottle
	return func(p *principal.Principal, ri *agent.RequestId, rsrs http.RequestStatusResponseStatus) error {
		time.Sleep(currentThrottling)
		currentThrottling *= time.Duration(backoffFactor * 100)
		currentThrottling = currentThrottling / 100
		return nil
	}
}

/**
 * Chain multiple polling strategy. This _chains_ the strategies, so if you pass in,
 * say, two throttling strategy of 1 second, it will result in a throttle of 2 seconds.
 * @param strategies A strategy list to chain.
 */
func chain(strategies ...PollStrategy) PollStrategy {
	strategy := func(p *principal.Principal, agent *agent.RequestId, status http.RequestStatusResponseStatus) error {
		for _, s := range strategies {
			err := s(p, agent, status)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return strategy
}
