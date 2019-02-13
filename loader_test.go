package konfig

import (
	"errors"
	"testing"
	time "time"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestLoaderHooksRun(t *testing.T) {
	t.Run(
		"run all hooks no error",
		func(t *testing.T) {
			var i int
			var loaderHooks = LoaderHooks{
				func(Store) error {
					i = i + 1
					return nil
				},
				func(Store) error {
					i = i + 2
					return nil
				},
				func(Store) error {
					i = i + 3
					return nil
				},
			}
			var err = loaderHooks.Run(Instance())
			require.Nil(t, err, "err should be nil")
			require.Equal(t, 6, i, "all hooks should have run")
		},
	)

	t.Run(
		"run one hook and error",
		func(t *testing.T) {
			var i int
			var loaderHooks = LoaderHooks{
				func(Store) error {
					i = i + 1
					return errors.New("err")
				},
				func(Store) error {
					i = i + 2
					return nil
				},
				func(Store) error {
					i = i + 3
					return nil
				},
			}
			var err = loaderHooks.Run(Instance())
			require.NotNil(t, err, "err should not be nil")
			require.Equal(t, 1, i, "one hook should have run")
		},
	)
}

func TestLoaderLoadRetry(t *testing.T) {
	var testCases = []struct {
		name  string
		err   bool
		build func(ctrl *gomock.Controller) *loaderWatcher
	}{
		{
			name: "success, no loader hooks, no retry",
			build: func(ctrl *gomock.Controller) *loaderWatcher {
				var mockW = NewMockWatcher(ctrl)
				var mockL = NewMockLoader(ctrl)
				mockL.EXPECT().Load(Values{}).Return(nil)

				var wl = &loaderWatcher{
					Watcher:     mockW,
					Loader:      mockL,
					loaderHooks: nil,
				}
				return wl
			},
		},
		{
			name: "success, no loader hooks, 1 retrty",
			build: func(ctrl *gomock.Controller) *loaderWatcher {
				var mockW = NewMockWatcher(ctrl)
				var mockL = NewMockLoader(ctrl)
				gomock.InOrder(
					mockL.EXPECT().Load(Values{}).Return(errors.New("")),
					mockL.EXPECT().RetryDelay().Return(1*time.Millisecond),
					mockL.EXPECT().MaxRetry().Return(1),
					mockL.EXPECT().Load(Values{}).Return(nil),
				)
				var wl = &loaderWatcher{
					Watcher:     mockW,
					Loader:      mockL,
					loaderHooks: nil,
				}
				return wl
			},
		},
		{
			name: "error, no loader hooks, 1 retry",
			err:  true,
			build: func(ctrl *gomock.Controller) *loaderWatcher {
				var mockW = NewMockWatcher(ctrl)
				var mockL = NewMockLoader(ctrl)
				gomock.InOrder(
					mockL.EXPECT().Load(Values{}).Return(errors.New("")),
					mockL.EXPECT().RetryDelay().Return(1*time.Millisecond),
					mockL.EXPECT().MaxRetry().Return(1),
					mockL.EXPECT().Load(Values{}).Return(errors.New("")),
					mockL.EXPECT().RetryDelay().Return(1*time.Millisecond),
					mockL.EXPECT().MaxRetry().Return(1),
				)
				var wl = &loaderWatcher{
					Watcher:     mockW,
					Loader:      mockL,
					loaderHooks: nil,
				}
				return wl
			},
		},
		{
			name: "success, 2 loader hooks, 1 retry",
			build: func(ctrl *gomock.Controller) *loaderWatcher {
				var mockW = NewMockWatcher(ctrl)
				var mockL = NewMockLoader(ctrl)
				gomock.InOrder(
					mockL.EXPECT().Load(Values{}).Return(errors.New("")),
					mockL.EXPECT().RetryDelay().Return(1*time.Millisecond),
					mockL.EXPECT().MaxRetry().Return(1),
					mockL.EXPECT().Load(Values{}).Return(nil),
				)
				var wl = &loaderWatcher{
					Watcher: mockW,
					Loader:  mockL,
					loaderHooks: LoaderHooks{
						func(Store) error {
							return nil
						},
						func(Store) error {
							return nil
						},
					},
				}
				return wl
			},
		},
		{
			name: "error, 2 loader hooks, 1 retry",
			err:  true,
			build: func(ctrl *gomock.Controller) *loaderWatcher {
				var mockW = NewMockWatcher(ctrl)
				var mockL = NewMockLoader(ctrl)
				gomock.InOrder(
					mockL.EXPECT().Load(Values{}).Return(errors.New("")),
					mockL.EXPECT().RetryDelay().Return(1*time.Millisecond),
					mockL.EXPECT().MaxRetry().Return(1),
					mockL.EXPECT().Load(Values{}).Return(nil),
				)
				var wl = &loaderWatcher{
					Watcher: mockW,
					Loader:  mockL,
					loaderHooks: LoaderHooks{
						func(Store) error {
							return nil
						},
						func(Store) error {
							return errors.New("")
						},
					},
				}
				return wl
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				var ctrl = gomock.NewController(t)
				defer ctrl.Finish()

				reset()
				var c = instance()
				c.cfg.NoExitOnError = true
				var err = c.loaderLoadRetry(testCase.build(ctrl), 0)
				if testCase.err {
					require.NotNil(t, err, "err should not be nil")
					return
				}
				require.Nil(t, err, "err should be nil")
			},
		)
	}
}

func TestLoaderLoadWatch(t *testing.T) {
	var testCases = []struct {
		name  string
		err   bool
		build func(ctrl *gomock.Controller) *loaderWatcher
	}{
		{
			name: "success, no errors",
			build: func(ctrl *gomock.Controller) *loaderWatcher {
				var mockW = NewMockWatcher(ctrl)
				var mockL = NewMockLoader(ctrl)

				mockL.EXPECT().Name().MinTimes(1).Return("test")
				mockL.EXPECT().Load(Values{}).Return(nil)
				mockW.EXPECT().Start().Return(nil)
				mockW.EXPECT().Watch().Return(nil)
				mockW.EXPECT().Done().Return(nil)

				var wl = &loaderWatcher{
					Watcher:     mockW,
					Loader:      mockL,
					loaderHooks: nil,
				}
				return wl
			},
		},
		{
			name: "success, errors load",
			err:  true,
			build: func(ctrl *gomock.Controller) *loaderWatcher {
				var mockW = NewMockWatcher(ctrl)
				var mockL = NewMockLoader(ctrl)

				mockL.EXPECT().Name().MinTimes(1).Return("test")
				mockL.EXPECT().Load(Values{}).Times(4).Return(errors.New(""))
				mockL.EXPECT().MaxRetry().Times(4).Return(3)
				mockL.EXPECT().RetryDelay().Times(4).Return(50 * time.Millisecond)
				mockL.EXPECT().StopOnFailure().Return(true)
				mockW.EXPECT().Close().MinTimes(1).Return(nil)

				var wl = &loaderWatcher{
					Watcher:     mockW,
					Loader:      mockL,
					loaderHooks: nil,
				}
				return wl
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				var ctrl = gomock.NewController(t)
				defer ctrl.Finish()

				reset()

				var c = New(&Config{
					Metrics: true,
				})

				c.RegisterLoaderWatcher(
					testCase.build(ctrl),
				)
				c.(*store).cfg.NoExitOnError = true

				var err = c.LoadWatch()

				if testCase.err {
					require.NotNil(t, err, "err should not be nil")
					return
				}

				require.Nil(t, err, "err should be nil")

				time.Sleep(300 * time.Millisecond)
			},
		)
	}
}
