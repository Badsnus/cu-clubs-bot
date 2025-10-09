package calendar

import (
	"bytes"
	"fmt"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"

	ics "github.com/arran4/golang-ical"
)

// ExportEventToICS converts a single event into an iCalendar (.ics) format.
// It creates a calendar, sets its properties, and adds the event to the calendar.
// Each event is assigned a unique identifier and properties such as creation time,
// start and end times, summary, description, location, status, transparency, and
// classification. Additionally, reminders are added for one day and one hour before
// the event. The function returns the serialized iCalendar data as a byte slice or
// an error if serialization fails.
func ExportEventToICS(event entity.Event) ([]byte, error) {
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetProductId("-//CU Clubs Bot//EN")
	cal.SetVersion("2.0")
	cal.SetCalscale("GREGORIAN")

	// Создаем уникальный идентификатор события
	uid := fmt.Sprintf("%s@cu-clubs-bot", event.ID)
	e := cal.AddEvent(uid)

	// Устанавливаем время создания и изменения события
	e.SetDtStampTime(event.CreatedAt)
	e.SetCreatedTime(event.CreatedAt)
	e.SetModifiedAt(event.UpdatedAt)

	// Устанавливаем время начала с указанием временной зоны
	e.SetStartAt(event.StartTime)

	// Установка времени окончания
	if !event.EndTime.IsZero() {
		e.SetEndAt(event.EndTime)
	} else {
		// Если время окончания не указано, устанавливаем продолжительность 1 час
		e.SetEndAt(event.StartTime.Add(1 * time.Hour))
	}

	// Устанавливаем основные свойства события
	e.SetSummary(event.Name)
	e.SetDescription(event.Description)
	e.SetLocation(event.Location)

	// Добавляем статус события
	e.SetStatus(ics.ObjectStatusConfirmed)

	// Добавляем прозрачность (показывает, занято ли время в календаре)
	e.SetTimeTransparency(ics.TransparencyOpaque)

	// Добавляем класс доступности (публичное)
	e.SetClass(ics.ClassificationPublic)

	// Добавляем последовательность (для синхронизации)
	e.SetSequence(0)

	// Добавляем напоминание за день до события
	dayAlarm := e.AddAlarm()
	dayAlarm.SetAction(ics.ActionDisplay)
	dayAlarm.AddProperty("TRIGGER;VALUE=DURATION", "-P1D")
	dayAlarm.SetDescription(fmt.Sprintf("Напоминание: %s (завтра)", event.Name))

	// Добавляем напоминание за час до события
	hourAlarm := e.AddAlarm()
	hourAlarm.SetAction(ics.ActionDisplay)
	hourAlarm.AddProperty("TRIGGER;VALUE=DURATION", "-PT1H")
	hourAlarm.SetDescription(fmt.Sprintf("Напоминание: %s (через час)", event.Name))

	var buf bytes.Buffer
	err := cal.SerializeTo(&buf)
	if err != nil {
		return nil, fmt.Errorf("error serializing calendar: %w", err)
	}

	return buf.Bytes(), nil
}
