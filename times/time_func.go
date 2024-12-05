package times

import "time"

// IntegerSecond 取整秒
func IntegerSecond(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, t.Hour(), t.Minute(), t.Second(), 0, t.Location())
}

// IntegerMinute 取整分
func IntegerMinute(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, t.Hour(), t.Minute(), 0, 0, t.Location())
}

// IntegerHour 取整点
func IntegerHour(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, t.Hour(), 0, 0, 0, t.Location())
}

// IntegerDay 取整天
func IntegerDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// IntegerWeek 取整周,周一
func IntegerWeek(t time.Time) time.Time {
	year, month, day := t.Date()
	week := int(t.Weekday())
	if week == 0 {
		week = 7
	}
	return time.Date(year, month, day+1-week, 0, 0, 0, 0, t.Location())
}

// IntegerMonth 取整月
func IntegerMonth(t time.Time) time.Time {
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}

// IntegerQuarter 取整季度
func IntegerQuarter(t time.Time) time.Time {
	year := t.Year()
	return time.Date(year, Month(Quarter(t)*3-2), 1, 0, 0, 0, 0, t.Location())
}

// IntegerYear 取整年
func IntegerYear(t time.Time) time.Time {
	year := t.Year()
	return time.Date(year, 1, 1, 0, 0, 0, 0, t.Location())
}

// Quarter 获取季度
func Quarter(t time.Time) int {
	return (int(t.Month())-1)/3 + 1
}
