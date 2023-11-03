package memorycache

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/lxzan/memorycache/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestMemoryCache(t *testing.T) {
	var as = assert.New(t)

	t.Run("", func(t *testing.T) {
		var db = New(WithInterval(10*time.Millisecond, 10*time.Millisecond), WithBucketNum(1))
		db.Set("a", 1, 100*time.Millisecond)
		db.Set("b", 1, 300*time.Millisecond)
		db.Set("c", 1, 500*time.Millisecond)
		db.Set("d", 1, 700*time.Millisecond)
		db.Set("e", 1, 900*time.Millisecond)
		db.Set("c", 1, time.Millisecond)

		time.Sleep(200 * time.Millisecond)
		as.ElementsMatch(db.Keys(""), []string{"b", "d", "e"})
	})

	t.Run("", func(t *testing.T) {
		var db = New(WithInterval(10*time.Millisecond, 10*time.Millisecond))
		db.Set("a", 1, 100*time.Millisecond)
		db.Set("b", 1, 200*time.Millisecond)
		db.Set("c", 1, 500*time.Millisecond)
		db.Set("d", 1, 700*time.Millisecond)
		db.Set("e", 1, 2900*time.Millisecond)
		db.Set("a", 1, 400*time.Millisecond)

		time.Sleep(300 * time.Millisecond)
		as.ElementsMatch(db.Keys(""), []string{"a", "c", "d", "e"})
	})

	t.Run("", func(t *testing.T) {
		var db = New(WithInterval(10*time.Millisecond, 10*time.Millisecond))
		db.Set("a", 1, 100*time.Millisecond)
		db.Set("b", 1, 200*time.Millisecond)
		db.Set("c", 1, 400*time.Millisecond)
		db.Set("d", 1, 700*time.Millisecond)
		db.Set("d", 1, 400*time.Millisecond)

		time.Sleep(500 * time.Millisecond)
		as.Equal(0, len(db.Keys("")))
	})

	t.Run("batch", func(t *testing.T) {
		var count = 1000
		var mc = New(
			WithInterval(10*time.Millisecond, 10*time.Millisecond),
			WithBucketNum(1),
		)
		var m1 = make(map[string]int)
		var m2 = make(map[string]int64)
		for i := 0; i < count; i++ {
			key := string(utils.AlphabetNumeric.Generate(16))
			exp := time.Duration(rand.Intn(10)+1) * 100 * time.Millisecond
			mc.Set(key, i, exp)
			m1[key] = i
			m2[key] = mc.getExp(exp)
		}

		time.Sleep(500 * time.Millisecond)
		for k, v := range m1 {
			result, ok := mc.Get(k)
			if ts := time.Now().UnixMilli(); ts > m2[k] {
				if ts-m2[k] >= 10 {
					as.False(ok)
				}
				continue
			}

			as.True(ok)
			as.Equal(result.(int), v)
		}

		var wg = &sync.WaitGroup{}
		wg.Add(1)
		result, exist := mc.GetOrCreateWithCallback(string(utils.AlphabetNumeric.Generate(16)), "x", 500*time.Millisecond, func(ele *Element, reason Reason) {
			as.Equal(reason, ReasonExpired)
			as.Equal(ele.Value.(string), "x")
			wg.Done()
		})
		as.False(exist)
		as.Equal(result.(string), "x")
		wg.Wait()
	})

	t.Run("expire", func(t *testing.T) {
		var mc = New(
			WithBucketNum(1),
			WithMaxKeysDeleted(3),
			WithInterval(50*time.Millisecond, 100*time.Millisecond),
		)
		mc.Set("a", 1, 150*time.Millisecond)
		mc.Set("b", 1, 150*time.Millisecond)
		mc.Set("c", 1, 150*time.Millisecond)
		time.Sleep(200 * time.Millisecond)
	})
}

func TestMemoryCache_Set(t *testing.T) {
	t.Run("", func(t *testing.T) {
		var list []string
		var count = 10000
		var mc = New(WithInterval(100*time.Millisecond, 100*time.Millisecond))
		mc.Clear()
		for i := 0; i < count; i++ {
			key := string(utils.AlphabetNumeric.Generate(8))
			exp := rand.Intn(1000)
			if exp == 0 {
				list = append(list, key)
			}
			mc.Set(key, 1, time.Duration(exp)*time.Millisecond)
		}
		for i := 0; i < count; i++ {
			key := string(utils.AlphabetNumeric.Generate(8))
			list = append(list, key)
			exp := rand.Intn(1000) + 3000
			mc.Set(key, 1, time.Duration(exp)*time.Millisecond)
		}
		time.Sleep(1100 * time.Millisecond)
		assert.ElementsMatch(t, utils.Uniq(list), mc.Keys(""))
	})

	t.Run("overflow", func(t *testing.T) {
		var mc = New(
			WithBucketNum(1),
			WithBucketSize(0, 2),
		)
		mc.Set("ming", 1, 3*time.Hour)
		mc.Set("hong", 1, 1*time.Hour)
		mc.Set("feng", 1, 2*time.Hour)
		assert.ElementsMatch(t, mc.Keys(""), []string{"ming", "feng"})
	})
}

func TestMemoryCache_Get(t *testing.T) {
	t.Run("", func(t *testing.T) {
		var list0 []string
		var list1 []string
		var count = 10000
		var mc = New(WithInterval(100*time.Millisecond, 100*time.Millisecond))
		for i := 0; i < count; i++ {
			key := string(utils.AlphabetNumeric.Generate(8))
			exp := rand.Intn(1000)
			if exp == 0 {
				list1 = append(list1, key)
			} else {
				list0 = append(list0, key)
			}
			mc.Set(key, 1, time.Duration(exp)*time.Millisecond)
		}
		for i := 0; i < count; i++ {
			key := string(utils.AlphabetNumeric.Generate(8))
			list1 = append(list1, key)
			exp := rand.Intn(1000) + 3000
			mc.Set(key, 1, time.Duration(exp)*time.Millisecond)
		}
		time.Sleep(1100 * time.Millisecond)

		for _, item := range list0 {
			_, ok := mc.Get(item)
			assert.False(t, ok)
		}
		for _, item := range list1 {
			_, ok := mc.Get(item)
			assert.True(t, ok)
		}
	})

	t.Run("expire", func(t *testing.T) {
		var mc = New(
			WithInterval(10*time.Second, 10*time.Second),
		)

		var wg = &sync.WaitGroup{}
		wg.Add(1)

		mc.SetWithCallback("ming", 128, 10*time.Millisecond, func(ele *Element, reason Reason) {
			assert.Equal(t, reason, ReasonExpired)
			assert.Equal(t, ele.Value.(int), 128)
			wg.Done()
		})

		time.Sleep(2 * time.Second)
		v, ok := mc.Get("ming")
		assert.False(t, ok)
		assert.Nil(t, v)
		wg.Wait()
	})
}

func TestMemoryCache_GetWithTTL(t *testing.T) {
	var list []string
	var count = 10000
	var mc = New(WithInterval(100*time.Millisecond, 100*time.Millisecond))
	for i := 0; i < count; i++ {
		key := string(utils.AlphabetNumeric.Generate(8))
		exp := rand.Intn(1000) + 200
		list = append(list, key)
		mc.Set(key, 1, time.Duration(exp)*time.Millisecond)
	}
	var keys = mc.Keys("")
	for _, key := range keys {
		mc.GetWithTTL(key, 2*time.Second)
	}

	time.Sleep(1100 * time.Millisecond)

	for _, item := range list {
		_, ok := mc.Get(item)
		assert.True(t, ok)
	}

	mc.Delete(list[0])
	_, ok := mc.GetWithTTL(list[0], -1)
	assert.False(t, ok)
}

func TestMemoryCache_Delete(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		var count = 10000
		var mc = New(WithInterval(100*time.Millisecond, 100*time.Millisecond))
		for i := 0; i < count; i++ {
			key := string(utils.AlphabetNumeric.Generate(8))
			exp := rand.Intn(1000) + 200
			mc.Set(key, 1, time.Duration(exp)*time.Millisecond)
		}

		var keys = mc.Keys("")
		for i := 0; i < 100; i++ {
			deleted := mc.Delete(keys[i])
			assert.True(t, deleted)

			key := string(utils.AlphabetNumeric.Generate(8))
			deleted = mc.Delete(key)
			assert.False(t, deleted)
		}
		assert.Equal(t, mc.Len(), count-100)
	})

	t.Run("2", func(t *testing.T) {
		var mc = New()
		var wg = &sync.WaitGroup{}
		wg.Add(1)
		mc.SetWithCallback("ming", 1, -1, func(ele *Element, reason Reason) {
			assert.Equal(t, reason, ReasonDeleted)
			wg.Done()
		})
		mc.SetWithCallback("ting", 2, -1, func(ele *Element, reason Reason) {
			wg.Done()
		})
		go mc.Delete("ming")
		wg.Wait()
	})

	t.Run("3", func(t *testing.T) {
		var mc = New()
		var wg = &sync.WaitGroup{}
		wg.Add(1)
		mc.GetOrCreateWithCallback("ming", 1, -1, func(ele *Element, reason Reason) {
			assert.Equal(t, reason, ReasonDeleted)
			wg.Done()
		})
		mc.GetOrCreateWithCallback("ting", 2, -1, func(ele *Element, reason Reason) {
			wg.Done()
		})
		go mc.Delete("ting")
		wg.Wait()
	})
}

func TestMaxCap(t *testing.T) {
	var mc = New(
		WithBucketNum(1),
		WithBucketSize(10, 100),
		WithInterval(100*time.Millisecond, 100*time.Millisecond),
	)

	var wg = &sync.WaitGroup{}
	wg.Add(900)
	for i := 0; i < 1000; i++ {
		key := string(utils.AlphabetNumeric.Generate(16))
		mc.SetWithCallback(key, 1, -1, func(ele *Element, reason Reason) {
			assert.Equal(t, reason, ReasonOverflow)
			wg.Done()
		})
	}
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, mc.Len(), 100)
	wg.Wait()
}

func TestMemoryCache_SetWithCallback(t *testing.T) {
	var as = assert.New(t)
	var count = 1000
	var mc = New(
		WithBucketNum(16),
		WithInterval(10*time.Millisecond, 100*time.Millisecond),
	)
	defer mc.Clear()

	var wg = &sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		key := string(utils.AlphabetNumeric.Generate(16))
		exp := time.Duration(rand.Intn(1000)+10) * time.Millisecond
		mc.SetWithCallback(key, i, exp, func(ele *Element, reason Reason) {
			as.True(time.Now().UnixMilli() > ele.ExpireAt)
			as.Equal(reason, ReasonExpired)
			wg.Done()
		})
	}
	wg.Wait()
}

func TestMemoryCache_GetOrCreate(t *testing.T) {

	var count = 1000
	var mc = New(
		WithBucketNum(16),
		WithInterval(10*time.Millisecond, 100*time.Millisecond),
	)
	defer mc.Clear()

	for i := 0; i < count; i++ {
		key := string(utils.AlphabetNumeric.Generate(16))
		exp := time.Duration(rand.Intn(1000)+10) * time.Millisecond
		mc.GetOrCreate(key, i, exp)
	}
}

func TestMemoryCache_GetOrCreateWithCallback(t *testing.T) {
	var as = assert.New(t)

	t.Run("", func(t *testing.T) {
		var count = 1000
		var mc = New(
			WithBucketNum(16),
			WithInterval(10*time.Millisecond, 100*time.Millisecond),
		)
		defer mc.Clear()

		var wg = &sync.WaitGroup{}
		wg.Add(count)
		for i := 0; i < count; i++ {
			key := string(utils.AlphabetNumeric.Generate(16))
			exp := time.Duration(rand.Intn(1000)+10) * time.Millisecond
			mc.GetOrCreateWithCallback(key, i, exp, func(ele *Element, reason Reason) {
				as.True(time.Now().UnixMilli() > ele.ExpireAt)
				as.Equal(reason, ReasonExpired)
				wg.Done()
			})
		}
		wg.Wait()
	})

	t.Run("exists", func(t *testing.T) {
		var mc = New()
		mc.Set("ming", 1, -1)
		v, exist := mc.GetOrCreateWithCallback("ming", 2, time.Second, func(ele *Element, reason Reason) {})
		as.True(exist)
		as.Equal(v.(int), 1)
	})

	t.Run("create", func(t *testing.T) {
		var mc = New(
			WithBucketNum(1),
			WithBucketSize(0, 1),
		)
		mc.Set("ming", 1, -1)
		v, exist := mc.GetOrCreateWithCallback("wang", 2, time.Second, func(ele *Element, reason Reason) {})
		as.False(exist)
		as.Equal(v.(int), 2)
		as.Equal(mc.Len(), 1)
	})
}

func TestMemoryCache_Stop(t *testing.T) {
	var mc = New()
	mc.Stop()
	mc.Stop()

	select {
	case <-mc.ctx.Done():
	default:
		t.Fail()
	}
}
