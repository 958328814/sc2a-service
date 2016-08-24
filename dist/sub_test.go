package dist

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListSubs(t *testing.T) {
	_, err := ListSubs()
	assert.NoError(t, err)
}

func TestSubscribe(t *testing.T) {
	sub, err := Subscribe(Sub{
		Name:  "Name",
		Email: "Email",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, sub.ID)
	assert.NotEmpty(t, sub.Date)

	list, err := ListSubs()
	assert.NoError(t, err)
	assert.True(t, len(list) >= 1)
	assert.True(t, reflect.DeepEqual(&list[len(list)-1], sub))

	assert.NoError(t, Unsubscribe(sub.ID))
}

func TestUpdateSubscriber(t *testing.T) {
	sub, err := Subscribe(Sub{
		Name:  "Name",
		Email: "Email",
	})
	assert.NoError(t, err)

	sub.Name = "NAME"
	sub.Email = "EMAIL"
	sub, err = UpdateSubscriber(*sub)
	assert.NoError(t, err)

	list, err := ListSubs()
	assert.NoError(t, err)
	assert.True(t, len(list) >= 1)
	assert.True(t, reflect.DeepEqual(&list[len(list)-1], sub))
}

func TestUnsubscribe(t *testing.T) {
	sub, err := Subscribe(Sub{
		Name:  "Name",
		Email: "Email",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, sub.ID)
	assert.NotEmpty(t, sub.Date)

	assert.NoError(t, Unsubscribe(sub.ID))

	list, err := ListSubs()
	assert.NoError(t, err)
	assert.True(t, len(list) == 0 || list[len(list)-1].ID != sub.ID)
}
