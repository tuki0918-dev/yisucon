package checker

import (
	"container/list"
	"net/http"
)

type Scenario struct {
	Actions *list.List
}

const (
	fakeLoginCount = 2
	pagingCount    = 5
)

func NewInitScenario(c *Checker) *Scenario {

	queue := list.New()

	queue.Init()

	queue.PushBack(NewAction(http.MethodGet, "InitialiCheck", c.InitialCheck))
	queue.PushBack(NewAction(http.MethodGet, "FaviconCheck", c.FaviconCheck))
	queue.PushBack(NewAction(http.MethodGet, "JSCheck", c.JSCheck))
	queue.PushBack(NewAction(http.MethodGet, "CSSCheck", c.CSSCheck))

	return &Scenario{queue}
}

func NewDefaultScenario(c *Checker) *Scenario {

	queue := list.New()

	queue.Init()

	actions := Actions{
		NewAction(http.MethodGet, "PageLoadCheck", c.PageLoadCheck),
		NewAction(http.MethodGet, "MyPageCheck", c.MyPageCheck),
		NewAction(http.MethodGet, "LoginPageCheck", c.LoginPageCheck),
		NewAction(http.MethodPost, "FakeLoginCheck", c.FakeLoginCheck),
		NewAction(http.MethodPost, "LoginCheck", c.LoginCheck),
		NewAction(http.MethodGet, "PagingCheck", c.PagingCheck),
		NewAction(http.MethodGet, "SelfPageCheck", c.SelfPageCheck),
		NewAction(http.MethodGet, "UnfollowButtonCheck", c.UnfollowButtonCheck),
		NewAction(http.MethodPost, "UnfollowCheck", c.UnfollowCheck),
		NewAction(http.MethodGet, "RemoveFromTopCheck", c.RemoveFromTopCheck),
		NewAction(http.MethodGet, "FollowButtonCheck", c.FollowButtonCheck),
		NewAction(http.MethodPost, "FollowCheck", c.FollowCheck),
		NewAction(http.MethodGet, "FollowerTweetCheck", c.FollowerTweetCheck),
		NewAction(http.MethodGet, "UnfollowButtonCheck", c.UnfollowButtonCheck),
		NewAction(http.MethodPost, "HashTagTweetCheck", c.HashTagTweetCheck),
		NewAction(http.MethodGet, "TweetCheck", c.TweetCheck),
		NewAction(http.MethodGet, "HashTagCheck", c.HashTagCheck),
		NewAction(http.MethodGet, "TweetSearchCheck", c.TweetSearchCheck),
		NewAction(http.MethodPost, "LogoutCheck", c.LogoutCheck),
	}

	for _, action := range actions {
		switch action.Name {
		case "FakeLoginCheck":
			for i := 0; i < fakeLoginCount; i++ {
				queue.PushBack(NewAction(http.MethodGet, "FaviconCheck", c.FaviconCheck))
				queue.PushBack(action)
				queue.PushBack(NewAction(http.MethodGet, "JSCheck", c.JSCheck))
				queue.PushBack(NewAction(http.MethodGet, "CSSCheck", c.CSSCheck))
			}
		case "PagingCheck":
			for i := 0; i < pagingCount; i++ {
				queue.PushBack(action)
			}
		default:
			queue.PushBack(NewAction(http.MethodGet, "FaviconCheck", c.FaviconCheck))
			queue.PushBack(action)
			queue.PushBack(NewAction(http.MethodGet, "JSCheck", c.JSCheck))
			queue.PushBack(NewAction(http.MethodGet, "CSSCheck", c.CSSCheck))
		}
	}

	return &Scenario{queue}
}

func (s *Scenario) Close() {
	s.Actions.Init()
	s.Actions = nil
}

func (s *Scenario) Pop() *Action {
	return s.Actions.Remove(s.Actions.Front()).(*Action)
}

func (s *Scenario) IsEmpty() bool {
	return s.Actions.Len() == 0
}
