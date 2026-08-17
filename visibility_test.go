package main

import (
	"context"
	"strings"
	"testing"
	"time"

	thingscloud "github.com/arthursoares/things-cloud-sdk"
)

func TestPendingTasksInInactiveContainersAreNotActive(t *testing.T) {
	now := time.Now().UTC().Truncate(24 * time.Hour)
	future := now.Add(48 * time.Hour)

	tests := []struct {
		name        string
		projectOpts []taskOption
		viaHeading  bool
	}{
		{
			name: "completed project",
			projectOpts: []taskOption{
				withStatus(thingscloud.TaskStatusCompleted),
			},
		},
		{
			name: "canceled project",
			projectOpts: []taskOption{
				withStatus(thingscloud.TaskStatusCanceled),
			},
		},
		{
			name: "trashed project",
			projectOpts: []taskOption{
				withTrashed(),
			},
		},
		{
			name: "heading in completed project",
			projectOpts: []taskOption{
				withStatus(thingscloud.TaskStatusCompleted),
			},
			viaHeading: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectOpts := append([]taskOption{
				withTitle("Archived Project"),
				withTaskType(thingscloud.TaskTypeProject),
			}, tt.projectOpts...)
			items := []thingscloud.Item{makeTaskItem("project-1", projectOpts...)}

			childOpts := []taskOption{
				withTitle("Hidden pending child"),
				withSchedule(thingscloud.TaskScheduleSomeday),
				withScheduledDate(now),
			}
			if tt.viaHeading {
				items = append(items, makeTaskItem("heading-1",
					withTitle("Section"),
					withTaskType(thingscloud.TaskTypeHeading),
					withParent("project-1"),
				))
				childOpts = append(childOpts, withActionGroup("heading-1"))
			} else {
				childOpts = append(childOpts, withParent("project-1"))
			}
			items = append(items, makeTaskItem("task-today", childOpts...))
			items = append(items, makeTaskItem("task-upcoming",
				withTitle("Hidden upcoming child"),
				withSchedule(thingscloud.TaskScheduleSomeday),
				withScheduledDate(future),
				withParent("project-1"),
			))

			fc := newFakeCloud("test@example.com", items...)
			defer fc.Close()
			tmcp := newTestThingsMCP(t, fc)

			for _, args := range []map[string]any{
				{},
				{"schedule": "today"},
				{"schedule": "upcoming"},
			} {
				result, err := tmcp.handleFindTasks(context.Background(), makeReq(args))
				if err != nil {
					t.Fatalf("find tasks: %v", err)
				}
				assertNotError(t, result)
				if tasks := resultJSON[[]TaskOutput](t, result); len(tasks) != 0 {
					t.Fatalf("inactive descendants leaked from find_tasks(%v): %#v", args, tasks)
				}
			}

			type overviewResult struct {
				TodayTasks    []TaskOutput `json:"todayTasks"`
				UpcomingTasks []TaskOutput `json:"upcomingTasks"`
			}
			overview, err := tmcp.handleOverview(context.Background(), makeReq(map[string]any{"lookahead_days": 7}))
			if err != nil {
				t.Fatalf("overview: %v", err)
			}
			assertNotError(t, overview)
			got := resultJSON[overviewResult](t, overview)
			if len(got.TodayTasks) != 0 || len(got.UpcomingTasks) != 0 {
				t.Fatalf("inactive descendants leaked from overview: %#v", got)
			}

			shown, err := tmcp.handleShowTask(context.Background(), makeReq(map[string]any{"uuid": "task-today"}))
			if err != nil {
				t.Fatalf("show task: %v", err)
			}
			assertNotError(t, shown)
			if task := resultJSON[TaskDetailOutput](t, shown); task.Status != "pending" {
				t.Fatalf("show_task must preserve raw child status, got %q", task.Status)
			}
		})
	}
}

func TestTaskWritesRejectInactiveContainers(t *testing.T) {
	tests := []struct {
		name        string
		projectOpts []taskOption
		request     map[string]any
		wantError   string
	}{
		{
			name:        "completed project",
			projectOpts: []taskOption{withStatus(thingscloud.TaskStatusCompleted)},
			request:     map[string]any{"title": "Must not be created", "project_uuid": "project-1"},
			wantError:   "project is completed",
		},
		{
			name:        "canceled project",
			projectOpts: []taskOption{withStatus(thingscloud.TaskStatusCanceled)},
			request:     map[string]any{"title": "Must not be created", "project_uuid": "project-1"},
			wantError:   "project is canceled",
		},
		{
			name:        "trashed project",
			projectOpts: []taskOption{withTrashed()},
			request:     map[string]any{"title": "Must not be created", "project_uuid": "project-1"},
			wantError:   "project is in trash",
		},
		{
			name:        "heading in completed project",
			projectOpts: []taskOption{withStatus(thingscloud.TaskStatusCompleted)},
			request:     map[string]any{"title": "Must not be created", "heading_uuid": "heading-1"},
			wantError:   "heading is inside an inactive project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectOpts := append([]taskOption{
				withTitle("Inactive Project"),
				withTaskType(thingscloud.TaskTypeProject),
			}, tt.projectOpts...)
			fc := newFakeCloud("test@example.com",
				makeTaskItem("project-1", projectOpts...),
				makeTaskItem("heading-1",
					withTitle("Section"),
					withTaskType(thingscloud.TaskTypeHeading),
					withParent("project-1"),
				),
			)
			defer fc.Close()
			tmcp := newTestThingsMCP(t, fc)

			result, err := tmcp.handleCreateTask(context.Background(), makeReq(tt.request))
			if err != nil {
				t.Fatalf("create task: %v", err)
			}
			assertIsError(t, result)
			if text := resultText(t, result); !strings.Contains(text, tt.wantError) {
				t.Fatalf("error %q does not contain %q", text, tt.wantError)
			}
			if commits := fc.getCommitLog(); len(commits) != 0 {
				t.Fatalf("rejected write posted %d commits", len(commits))
			}
		})
	}
}

func TestOverdueUpcomingTaskHasConsistentTodaySchedule(t *testing.T) {
	past := time.Now().UTC().Add(-24 * time.Hour)
	fc := newFakeCloud("test@example.com", makeTaskItem("task-1",
		withTitle("Overdue task"),
		withSchedule(thingscloud.TaskScheduleSomeday),
		withScheduledDate(past),
	))
	defer fc.Close()
	tmcp := newTestThingsMCP(t, fc)

	today, err := tmcp.handleFindTasks(context.Background(), makeReq(map[string]any{"schedule": "today"}))
	if err != nil {
		t.Fatalf("today query: %v", err)
	}
	assertNotError(t, today)
	tasks := resultJSON[[]TaskOutput](t, today)
	if len(tasks) != 1 || tasks[0].Schedule != "today" {
		t.Fatalf("today query returned inconsistent task: %#v", tasks)
	}

	upcoming, err := tmcp.handleFindTasks(context.Background(), makeReq(map[string]any{"schedule": "upcoming"}))
	if err != nil {
		t.Fatalf("upcoming query: %v", err)
	}
	assertNotError(t, upcoming)
	if tasks := resultJSON[[]TaskOutput](t, upcoming); len(tasks) != 0 {
		t.Fatalf("overdue task leaked into upcoming: %#v", tasks)
	}
}
