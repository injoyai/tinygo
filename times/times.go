package times

import "time"

var (
	//Offset 系统时间和网络偏移量
	Offset = time.Duration(0)
)

type (
	Duration = time.Duration
	Month    = time.Month
)

const (
	FormatDefault = "2006-01-02 15:04:05"
	FormatDate    = "2006-01-02"
	FormatTime    = "15:04:05"

	Nanosecond  = time.Nanosecond
	Microsecond = time.Microsecond
	Millisecond = time.Millisecond
	Second      = time.Second
	Minute      = time.Minute
	Hour        = time.Hour
	Day         = Hour * 24
	Week        = Day * 7
)

type Time struct {
	time.Time
}

// Now 取当前时间
func Now() Time {
	return Time{
		Time: time.Now().Add(Offset),
	}
}

// New 新建时间,秒
func New(sec int64) Time {
	return NewSec(sec)
}

func NewSec(sec int64) Time {
	return NewNano(sec * 1e9)
}

func NewMill(mill int64) Time {
	return NewNano(mill * 1e6)
}

func NewMicro(micro int64) Time {
	return NewNano(micro * 1e3)
}

func NewNano(nano int64) Time {
	return Time{
		Time: time.Unix(nano/1e9, nano%1e9),
	}
}

// Parse string类型转time
func Parse(format, value string) (Time, error) {
	t, err := time.Parse(format, value)
	return Time{Time: t}, err
}

// ParseDefault string类型转time
func ParseDefault(s string) (Time, error) {
	t, err := time.Parse(FormatDefault, s)
	return Time{Time: t}, err
}

func (this Time) String() string {
	return this.Format(FormatDefault)
}

// Date 年月日秒
func (this Time) Date() (int, int, int, int) {
	year, month, day := this.Time.Date()
	return year, int(month), day, this.Second()
}

// Quarter 1-4季,当年
func (this Time) Quarter() int {
	return Quarter(this.Time)
}

func (this Time) Duration() time.Duration {
	return time.Duration(this.UnixNano())
}

//********************************************分界线**************************************************//

func (this Time) Add(d time.Duration) Time {
	this.Time = this.Time.Add(d)
	return this
}

// AddSec 秒,加减
func (this Time) AddSec(sec int) Time {
	return this.Add(time.Second * time.Duration(sec))
}

// AddMin 分,加减
func (this Time) AddMin(minute int) Time {
	return this.Add(time.Minute * time.Duration(minute))
}

// AddHour 小时加减
func (this Time) AddHour(hour int) Time {
	return this.Add(time.Hour * time.Duration(hour))
}

// AddDay 天,加减
func (this Time) AddDay(day int) Time {
	this.Time = this.AddDate(0, 0, day)
	return this
}

// AddWeek 周,加减
func (this Time) AddWeek(week int) Time {
	this.Time = this.AddDate(0, 0, week*7)
	return this
}

// AddMonth 月,加减
func (this Time) AddMonth(month int) Time {
	this.Time = this.AddDate(0, month, 0)
	return this
}

// AddYear 年,加减
func (this Time) AddYear(year int) Time {
	this.Time = this.AddDate(year, 0, 0)
	return this
}

//************************************************分界线**********************************************//

// IntegerSecond 取整秒
func (this Time) IntegerSecond() Time {
	this.Time = IntegerSecond(this.Time)
	return this
}

// IntegerMinute 取整分
func (this Time) IntegerMinute() Time {
	this.Time = IntegerMinute(this.Time)
	return this
}

// IntegerHour 取整点
func (this Time) IntegerHour() Time {
	this.Time = IntegerHour(this.Time)
	return this
}

// IntegerDay 取整天
func (this Time) IntegerDay() Time {
	this.Time = IntegerDay(this.Time)
	return this
}

// IntegerWeek 取整周,周一
func (this Time) IntegerWeek() Time {
	this.Time = IntegerWeek(this.Time)
	return this
}

// IntegerMonth 取整月
func (this Time) IntegerMonth() Time {
	this.Time = IntegerMonth(this.Time)
	return this
}

func (this Time) IntegerQuarter() Time {
	this.Time = IntegerQuarter(this.Time)
	return this
}

// IntegerYear 取整年
func (this Time) IntegerYear() Time {
	this.Time = IntegerYear(this.Time)
	return this
}
