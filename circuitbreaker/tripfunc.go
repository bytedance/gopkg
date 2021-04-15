// Copyright 2021 ByteDance Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package circuitbreaker

import "time"

// TripFunc is a function called by a breaker when error appear and
// determines whether the breaker should trip.
type TripFunc func(Metricer) bool

// TripFuncWithKey returns a TripFunc according to the key.
type TripFuncWithKey func(string) TripFunc

// ThresholdTripFunc .
func ThresholdTripFunc(threshold int64) TripFunc {
	return func(m Metricer) bool {
		return m.Failures()+m.Timeouts() >= threshold
	}
}

// ConsecutiveTripFunc .
func ConsecutiveTripFunc(threshold int64) TripFunc {
	return func(m Metricer) bool {
		return m.ConseErrors() >= threshold
	}
}

// RateTripFunc .
func RateTripFunc(rate float64, minSamples int64) TripFunc {
	return func(m Metricer) bool {
		samples := m.Samples()
		return samples >= minSamples && m.ErrorRate() >= rate
	}
}

// ConsecutiveTripFuncV2 根据传入的参数进行判断，分别采用以下三种策略：
// 1. 当样本数 >= minSamples 且 错误率 >= rate
// 2. 当样本数 >= durationSamples 且 连续出错时长 >= duration
// 3. 当连续错误数 >= conseErrors
// 以上三种策略成立任何一种就打开熔断器。
func ConsecutiveTripFuncV2(rate float64, minSamples int64, duration time.Duration, durationSamples, conseErrors int64) TripFunc {
	return func(m Metricer) bool {
		samples := m.Samples()
		// 基于统计
		if samples >= minSamples && m.ErrorRate() >= rate {
			return true
		}
		// 基于连续时长
		if duration > 0 && m.ConseErrors() >= durationSamples && m.ConseTime() >= duration {
			return true
		}
		// 基于连续错误数
		if conseErrors > 0 && m.ConseErrors() >= conseErrors {
			return true
		}
		return false
	}
}
