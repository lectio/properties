package properties

import (
	"context"
	"github.com/araddon/dateparse"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const validFrontMatter = `
---
description: test description
number: 221
flag: true
date: 2006-01-02T15:04:05Z07:00
---
test body
`

const noFrontMatter = `test body without front matter`

const invalidFrontMatter1 = `
---
description: test description

test body
`

type PropertiesSuite struct {
	suite.Suite
	factory Factory
}

func (suite *PropertiesSuite) SetupSuite() {
	suite.factory = ThePropertiesFactory
}

func (suite *PropertiesSuite) TearDownSuite() {
}

func (suite *PropertiesSuite) TestMutableProperties() {
	ctx := context.Background()
	props := suite.factory.EmptyMutable(ctx)
	suite.NotNil(props, "Ensure initialization")
	suite.Equal(uint(0), props.Size(ctx), "Should be zero")

	prop, ok, err := props.AddAny(ctx, "custom", suite)
	suite.False(ok, "Should not have been created")
	suite.NotNil(err, "Should have gotten an error")

	prop, ok, err = props.AddAny(ctx, "text", "Test text")
	suite.True(ok, "Should have been created")
	suite.IsType(&DefaultTextProperty{}, prop, "Should have been created")

	prop, ok, err = props.AddAny(ctx, "number", 100)
	prop, ok, err = props.AddAny(ctx, "flag", true)
	prop, ok, err = props.AddAny(ctx, "date", time.Now())
}

func (suite *PropertiesSuite) TestNoFrontMatter() {
	ctx := context.Background()
	bodyBytes, props, count, err := suite.factory.MutableFromFrontMatter(ctx, []byte(noFrontMatter), false)
	body := string(bodyBytes)
	suite.Nil(err, "Shouldn't have any errors")
	suite.Nil(props, "Should not be initialized, there is no front matter")
	suite.Equal(uint(0), count, "Should not have any front matter")
	suite.Equal(body, noFrontMatter)
}

func (suite *PropertiesSuite) TestValidFrontMatter() {
	ctx := context.Background()
	bodyBytes, props, count, err := suite.factory.MutableFromFrontMatter(ctx, []byte(validFrontMatter), false)
	body := string(bodyBytes)

	suite.Nil(err, "Shouldn't have any errors")
	suite.NotNil(props, "Should be initialized")
	suite.Equal(uint(4), count, "Should have four items")
	suite.Equal(body, "test body")

	prop, _ := props.Named(ctx, "description")
	suite.Equal("test description", prop.AnyValue(ctx))

	prop, _ = props.Named(ctx, "number")
	suite.Equal(int64(221), prop.AnyValue(ctx))

	prop, _ = props.Named(ctx, "flag")
	suite.Equal(true, prop.AnyValue(ctx))

	prop, _ = props.Named(ctx, "date")
	suite.Equal("2006-01-02T15:04:05Z07:00", prop.AnyValue(ctx))
}

func (suite *PropertiesSuite) TestValidSmartParsedFrontMatter() {
	ctx := context.Background()
	bodyBytes, props, count, err := suite.factory.MutableFromFrontMatter(ctx, []byte(validFrontMatter), true)
	body := string(bodyBytes)

	suite.Nil(err, "Shouldn't have any errors")
	suite.NotNil(props, "Should be initialized")
	suite.Equal(uint(4), count, "Should have four items")
	suite.Equal(body, "test body")

	prop, _ := props.Named(ctx, "description")
	suite.Equal("test description", prop.AnyValue(ctx))

	prop, _ = props.Named(ctx, "number")
	suite.Equal(int64(221), prop.AnyValue(ctx))

	prop, _ = props.Named(ctx, "flag")
	suite.Equal(true, prop.AnyValue(ctx))

	prop, _ = props.Named(ctx, "date")
	date, _ := dateparse.ParseAny("2006-01-02T15:04:05Z07:00")
	suite.Equal(date, prop.AnyValue(ctx))
}

func (suite *PropertiesSuite) TestInvalidFrontMatter() {
	ctx := context.Background()
	bodyBytes, props, count, err := suite.factory.MutableFromFrontMatter(ctx, []byte(invalidFrontMatter1), false)

	suite.EqualError(err, "Unexplained front matter parser error; insideFrontMatter: true, yamlStartIndex: 5, yamlEndIndex: 0")
	suite.Nil(props, "Should not be initialized")
	suite.Equal(uint(0), count, "Should not have any front matter")
	suite.Nil(bodyBytes, "Body should be empty")
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(PropertiesSuite))
}
