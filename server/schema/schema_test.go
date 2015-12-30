package schema_test

import (
	"encoding/json"
	"fmt"
	"github.com/SuperLimitBreak/socketServerMq/server/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnmarshalMessage(t *testing.T) {
	//Real data taken from even_map.json in displayTrigger
	payloads := []string{
		`{"deviceid": "test", "func": "audio.start", "src": "/assets/test_gurren.ogg", "target_selector": "#screen", "currentTime": 4.22}`,
		`{"deviceid": "main", "func": "video.precache", "src": "/assets/gurren_lagann.mp4"}`,
		`{"deviceid": "lights", "func": "DMXRendererLightTiming.start", "scene": "gurren_intro_piano", "bpm": 110.0}`,
		`{"deviceid": "subtitles", "func": "subtitles.load", "src": "/assets/gurren.srt", "play_on_load": false}`,
	}

	payloadKeys := []string{
		"test",
		"main",
		"lights",
		"subtitles",
	}

	jsonFmtStr := `{
		"action":"message",
		"data":[
			%v,
			%v,
			%v,
			%v
		]
	}`

	message := []byte(
		fmt.Sprintf(jsonFmtStr,
			payloads[0], payloads[1],
			payloads[2], payloads[3],
		),
	)

	var generic schema.GenericMessage
	err := json.Unmarshal(message, &generic)

	assert.Nil(t, err)
	assert.Equal(t, "message", schema.MESSAGE_ACTION)
	assert.Equal(t, schema.MESSAGE_ACTION, generic.Action)
	assert.True(t, generic.IsAction(schema.MESSAGE_ACTION))

	splitMessages, err := generic.Messages()

	assert.Nil(t, err)
	assert.Len(t, splitMessages, 4)

	for i, elm := range splitMessages {
		assert.EqualValues(t, []byte(payloads[i]), *elm.Data)

		keyPtr, err := elm.Key()

		assert.Nil(t, err)
		if assert.NotNil(t, keyPtr) {
			assert.Equal(t, payloadKeys[i], *keyPtr)
		}
	}

}
