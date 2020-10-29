//
// Copyright 2019 Lightbend Inc.
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

package crdt

import "github.com/cloudstateio/go-support/cloudstate/entity"

type Clock uint64

const (
	// The Default clock, uses the current system time as the clock value.
	Default Clock = iota

	// A Reverse clock, based on the system clock. Using this effectively
	// achieves First-Write-Wins semantics. This is susceptible to the
	// same clock skew problems as the default clock.
	Reverse

	// A custom clock.
	// The custom clock value is passed by using the customClockValue parameter on
	// the `SetWithClock` method. The value should be a domain specific monotonically
	// increasing value. For example, if the source of the value for this register
	// is a single device, that device may attach a sequence number to each update,
	// that sequence number can be used to guarantee that the register will converge
	// to the last update emitted by that device.
	Custom

	// CustomAutoIncrement is a custom clock, that automatically increments the
	// custom value if the local clock value is greater than it.
	//
	// This is like `Custom`, however if when performing the update in the proxy,
	// it's found that the clock value of the register is greater than the specified
	// clock value for the update, the proxy will instead use the current clock
	// value of the register plus one.
	//
	// This can guarantee that updates done on the same node will be causally
	// ordered (addressing problems caused by the system clock being adjusted),
	// but will not guarantee causal ordering for updates on different nodes,
	// since it's possible that an update on a different node has not yet been
	// replicated to this node.
	CustomAutoIncrement
)

func fromCrdtClock(clock entity.CrdtClock) Clock {
	switch clock {
	case entity.CrdtClock_DEFAULT:
		return Default
	case entity.CrdtClock_REVERSE:
		return Reverse
	case entity.CrdtClock_CUSTOM:
		return Custom
	case entity.CrdtClock_CUSTOM_AUTO_INCREMENT:
		return CustomAutoIncrement
	default:
		return Default
	}
}

func (c Clock) toCrdtClock() entity.CrdtClock {
	switch c {
	case Default:
		return entity.CrdtClock_DEFAULT
	case Reverse:
		return entity.CrdtClock_REVERSE
	case Custom:
		return entity.CrdtClock_CUSTOM
	case CustomAutoIncrement:
		return entity.CrdtClock_CUSTOM_AUTO_INCREMENT
	default:
		return entity.CrdtClock_DEFAULT
	}
}
