package jog

import (
	"fmt"
	"testing"
)

var (
	levelFromTest = []levelTest{
		{DEBUG, map[string]interface{}{"level": "debug"}},
		{INFO, map[string]interface{}{"level": "info"}},
		{WARNING, map[string]interface{}{"level": "warning"}},
		{ERROR, map[string]interface{}{"level": "error"}},
		{CRITICAL, map[string]interface{}{"level": "critical"}},

		// Should get INFO for all invalid values
		{INFO, map[string]interface{}{"level": "bob"}},
		{INFO, map[string]interface{}{"level": "DEBUG"}},
		{INFO, map[string]interface{}{"level": "Jack"}},
		{INFO, map[string]interface{}{"level": "    "}},
		{INFO, map[string]interface{}{"level": "-1"}},
		{INFO, map[string]interface{}{"level": "1"}},
		{INFO, map[string]interface{}{"level": 1}},
		{INFO, map[string]interface{}{"level": []string{"critical"}}},
		{INFO, map[string]interface{}{"level": nil}},
	}

	newMessageTest = []messageTest{
		{DEBUG, -1, "blah blah", "blah blah"},
		{INFO, 1, "blah blah", "blah blah"},
		{WARNING, 0, "blah blah", "blah blah"},
		{ERROR, 0, "blah blah", "blah blah"},
		{CRITICAL, 0, "blah blah", "blah blah"},

		{CRITICAL, 0, nil, nil},
		{CRITICAL, 0, dummy1{""}, "{}"},
		{CRITICAL, 0, dummy1{"blah blah"}, "{blah blah}"},
		{CRITICAL, 0, dummy2{dummy1{"blah blah"}}, "dummy2"},
	}

	writeTests = []writeTest{
		{INFO, "blah blah", "blah blah"},
		{INFO, `{blah"`, `{blah`},
		{INFO, `{"blah": "blah"`, `{"blah": "blah"`},
		// Test JSON conversion
		{INFO, `{"message": "blah blah"}`, map[string]interface{}{"message": "blah blah"}},
		// Ensure `level` is remove from source
		{ERROR, `{"message": "blah blah", "level": "error"}`, map[string]interface{}{"message": "blah blah"}},
		{CRITICAL, `{"level": "critical"}`, map[string]interface{}{}},
		// Test nested objects
		{DEBUG, `{"message": { "innerMessage": "blah" }, "level": "debug"}`, map[string]interface{}{"message": map[string]interface{}{"innerMessage": "blah"}}},
	}
)

type testLogger struct {
	message *Message
}

type levelTest struct {
	expected Level
	value    map[string]interface{}
}

type messageTest struct {
	level    Level
	depth    int
	message  interface{}
	expected interface{}
}

type writeTest struct {
	expectedLevel   Level
	message         string
	expectedMessage interface{}
}

type dummy1 struct {
	message string
}
type dummy2 struct {
	dummy1
}

func (d dummy2) String() string {
	return "dummy2"
}

func (l *testLogger) Log(m *Message) (int, error) {
	l.message = m
	return -1, nil
}

func TestLevelFrom(t *testing.T) {
	for _, v := range levelFromTest {
		l := levelFrom(v.value)
		if l != v.expected {
			t.Error("Expected", v.expected, "got", l)
		}
	}
}

func TestNewMessage(t *testing.T) {
	for _, v := range newMessageTest {
		m := newMessage(v.level, v.message, v.depth)
		if m.Data != v.expected {
			t.Errorf("Expected '%v' got '%v'", v.expected, m.Data)
		}
	}
}

func TestWrite(t *testing.T) {
	l := &testLogger{}
	j := New(l)

	for _, v := range writeTests {
		p := []byte(v.message)
		if _, err := j.Write(p); err != nil {
			t.Error("Failed to write message during writeTest", err)
		}

		// Check that we've got the right type, and that it's the expected value
		switch l.message.Data.(type) {
		case map[string]interface{}:
			if _, ok := v.expectedMessage.(map[string]interface{}); !ok {
				t.Errorf("Expected type %T got %T", v.expectedMessage, l.message.Data)
			}
			s1, s2 := fmt.Sprint(v.expectedMessage), fmt.Sprint(l.message.Data)
			if s1 != s2 {
				t.Error("Expected", s1, "got", s2)
			}
		}

		// Check that the level was correctly set
		if l.message.Level != v.expectedLevel {
			t.Errorf("Expected level %s got %s", v.expectedLevel, l.message.Level)
		}
	}
}
