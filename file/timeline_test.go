package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyTimeline(t *testing.T) {
	timeline := NewTimeline()
	assertPrevAndNext(t, timeline, TimelineState{}, TimelineState{})
}

func TestTimelineTransitionFrom(t *testing.T) {
	timeline := NewTimeline()
	s1 := TimelineState{Path: "f1", LineNum: 1}
	timeline.TransitionFrom(s1)
	assertPrevAndNext(t, timeline, s1, TimelineState{})

	s2 := TimelineState{Path: "f2", LineNum: 2}
	timeline.TransitionFrom(s2)
	assertPrevAndNext(t, timeline, s2, TimelineState{})
}

func TestTimelineMoveBackwardAndForward(t *testing.T) {
	timeline := NewTimeline()
	states := []TimelineState{
		{Path: "f1", LineNum: 1},
		{Path: "f2", LineNum: 2},
		{Path: "f3", LineNum: 3},
		{Path: "f4", LineNum: 4},
	}
	for _, s := range states {
		timeline.TransitionFrom(s)
	}

	// backward f5:5 -> f4:4
	s1 := TimelineState{Path: "f5", LineNum: 5}
	timeline.TransitionBackwardFrom(s1)
	assertPrevAndNext(t, timeline, states[2], s1)

	// backward f4:6 -> f3:3
	s2 := TimelineState{Path: "f4", LineNum: 6}
	timeline.TransitionBackwardFrom(s2)
	assertPrevAndNext(t, timeline, states[1], s2)

	// forward f3:7 -> f4:6
	s3 := TimelineState{Path: "f3", LineNum: 7}
	timeline.TransitionForwardFrom(s3)
	assertPrevAndNext(t, timeline, s3, s1)

	// forward f4:8 -> f5:5
	s4 := TimelineState{Path: "f4", LineNum: 8}
	timeline.TransitionForwardFrom(s4)
	assertPrevAndNext(t, timeline, s4, TimelineState{})

	// backward f5:9 -> f4:8
	s5 := TimelineState{Path: "f5", LineNum: 9}
	timeline.TransitionBackwardFrom(s5)
	assertPrevAndNext(t, timeline, s3, s5)
}

func TestTimelineMoveBackwardThenTransition(t *testing.T) {
	timeline := NewTimeline()
	states := []TimelineState{
		{Path: "f1", LineNum: 1},
		{Path: "f2", LineNum: 2},
		{Path: "f3", LineNum: 3},
		{Path: "f4", LineNum: 4},
	}
	for _, s := range states {
		timeline.TransitionFrom(s)
	}

	// backward f5:5 -> f4:4
	s1 := TimelineState{Path: "f5", LineNum: 5}
	timeline.TransitionBackwardFrom(s1)
	assertPrevAndNext(t, timeline, states[2], s1)

	// backward f4:6 -> f3:3
	s2 := TimelineState{Path: "f4", LineNum: 6}
	timeline.TransitionBackwardFrom(s2)
	assertPrevAndNext(t, timeline, states[1], s2)

	// transition f3:7 -> f6:8
	s3 := TimelineState{Path: "f3", LineNum: 7}
	timeline.TransitionFrom(s3)
	assertPrevAndNext(t, timeline, s3, TimelineState{})
}

func assertPrevAndNext(t *testing.T, timeline *Timeline, prev, next TimelineState) {
	assert.Equal(t, prev, timeline.PeekBackward())
	assert.Equal(t, next, timeline.PeekForward())
}
