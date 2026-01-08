package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
)

func TestCalendarUseCase_CreateCalendar(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockCalendarRepository(ctrl)
	uc := NewCalendarUseCase(mockRepo, nil, zap.NewNop())

	req := CreateCalendarRequest{Name: "Team", Type: domain.CalendarTypePersonal, OwnerID: 1}

	mockRepo.EXPECT().CreateCalendar(gomock.Any(), gomock.Any()).Return(nil)
	cal, err := uc.CreateCalendar(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "Team", cal.Name)

	mockRepo.EXPECT().CreateCalendar(gomock.Any(), gomock.Any()).Return(errors.New("db"))
	cal, err = uc.CreateCalendar(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, cal)
}

func TestCalendarUseCase_GetAndList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockCalendarRepository(ctrl)
	uc := NewCalendarUseCase(mockRepo, nil, zap.NewNop())

	cal := &domain.Calendar{ID: 1, Name: "Team"}

	mockRepo.EXPECT().GetCalendar(gomock.Any(), uint(1)).Return(cal, nil)
	res, err := uc.GetCalendar(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, cal, res)

	mockRepo.EXPECT().GetCalendar(gomock.Any(), uint(1)).Return(nil, errors.New("missing"))
	res, err = uc.GetCalendar(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, res)

	list := []domain.Calendar{*cal}
	mockRepo.EXPECT().ListCalendars(gomock.Any(), uint(1), 1, 10).Return(list, 1, nil)
	lres, err := uc.ListCalendars(context.Background(), ListCalendarsRequest{OwnerID: 1, Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, 1, lres.Total)

	mockRepo.EXPECT().ListCalendars(gomock.Any(), uint(1), 1, 10).Return(nil, 0, errors.New("db"))
	lres, err = uc.ListCalendars(context.Background(), ListCalendarsRequest{OwnerID: 1, Page: 1, PageSize: 10})
	assert.Error(t, err)
	assert.Nil(t, lres)
}

func TestCalendarUseCase_CreateEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockCalendarRepository(ctrl)
	uc := NewCalendarUseCase(mockRepo, nil, zap.NewNop())

	start := time.Now()
	end := start.Add(time.Hour)
	req := CreateEventRequest{CalendarID: 1, Title: "Meeting", StartTime: start, EndTime: end, Type: domain.EventTypeMeeting, CreatedByID: 1}

	// success
	mockRepo.EXPECT().GetCalendar(gomock.Any(), uint(1)).Return(&domain.Calendar{ID: 1}, nil)
	mockRepo.EXPECT().CreateEvent(gomock.Any(), gomock.Any()).Return(nil)
	evt, err := uc.CreateEvent(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "Meeting", evt.Title)

	// conflict
	req.CheckConflicts = true
	mockRepo.EXPECT().GetCalendar(gomock.Any(), uint(1)).Return(&domain.Calendar{ID: 1}, nil)
	mockRepo.EXPECT().GetEventsByTimeRange(gomock.Any(), uint(1), gomock.Any(), gomock.Any()).Return([]domain.Event{{ID: 2}}, nil)
	evt, err = uc.CreateEvent(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, evt)
}
