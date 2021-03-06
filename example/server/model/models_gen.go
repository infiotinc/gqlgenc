// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

type Book interface {
	IsBook()
}

type Media interface {
	IsMedia()
}

type Chatroom struct {
	Name     string        `json:"name"`
	Messages []*Message    `json:"messages"`
	Hash     *FooTypeHash1 `json:"hash"`
}

type ColoringBook struct {
	Title  string   `json:"title"`
	Colors []string `json:"colors"`
}

func (ColoringBook) IsBook() {}

type Cyclic1_1 struct {
	ID    string     `json:"id"`
	Child *Cyclic1_2 `json:"child"`
}

type Cyclic1_2 struct {
	ID    string     `json:"id"`
	Child *Cyclic1_1 `json:"child"`
}

type Cyclic2_1 struct {
	ID    string     `json:"id"`
	Child *Cyclic2_2 `json:"child"`
}

type Cyclic2_2 struct {
	ID    string     `json:"id"`
	Child *Cyclic2_1 `json:"child"`
}

type Image struct {
	Size int `json:"size"`
}

func (Image) IsMedia() {}

type InputIssue14 struct {
	Ids []string `json:"ids"`
}

type Issue8Payload struct {
	Foo1 *Issue8PayloadFoo `json:"foo1"`
	Foo2 *Issue8PayloadFoo `json:"foo2"`
}

type Issue8PayloadFoo struct {
	A *Issue8PayloadFooA `json:"a"`
}

type Issue8PayloadFooA struct {
	Aa string `json:"Aa"`
}

type Message struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	CreatedBy string    `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
}

type OptionalValue1 struct {
	Value *string `json:"value"`
}

type OptionalValue2 struct {
	Value *string `json:"value"`
}

type PostCreateInput struct {
	Text string `json:"text"`
}

type SomeExtraType struct {
	Child *SomeExtraTypeChild `json:"child"`
}

type SomeExtraTypeChild struct {
	Child *SomeExtraTypeChildChild `json:"child"`
}

type SomeExtraTypeChildChild struct {
	ID string `json:"id"`
}

type Textbook struct {
	Title   string   `json:"title"`
	Courses []string `json:"courses"`
}

func (Textbook) IsBook() {}

type UploadData struct {
	Size int `json:"size"`
}

type UploadFilesMap struct {
	Somefile *UploadData `json:"somefile"`
}

type UploadFilesMapInput struct {
	Somefile graphql.Upload `json:"somefile"`
}

type Video struct {
	Duration int `json:"duration"`
}

func (Video) IsMedia() {}

type Episode string

const (
	EpisodeNewhope Episode = "NEWHOPE"
	EpisodeEmpire  Episode = "EMPIRE"
	EpisodeJedi    Episode = "JEDI"
)

var AllEpisode = []Episode{
	EpisodeNewhope,
	EpisodeEmpire,
	EpisodeJedi,
}

func (e Episode) IsValid() bool {
	switch e {
	case EpisodeNewhope, EpisodeEmpire, EpisodeJedi:
		return true
	}
	return false
}

func (e Episode) String() string {
	return string(e)
}

func (e *Episode) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Episode(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Episode", str)
	}
	return nil
}

func (e Episode) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type FooTypeHash1 string

const (
	FooTypeHash1Hash1 FooTypeHash1 = "hash_1"
	FooTypeHash1Hash2 FooTypeHash1 = "hash_2"
)

var AllFooTypeHash1 = []FooTypeHash1{
	FooTypeHash1Hash1,
	FooTypeHash1Hash2,
}

func (e FooTypeHash1) IsValid() bool {
	switch e {
	case FooTypeHash1Hash1, FooTypeHash1Hash2:
		return true
	}
	return false
}

func (e FooTypeHash1) String() string {
	return string(e)
}

func (e *FooTypeHash1) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = FooTypeHash1(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid FooType_hash1", str)
	}
	return nil
}

func (e FooTypeHash1) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
