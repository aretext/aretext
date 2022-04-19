package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTodoTxtParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name:     "description only",
			text:     "Fix a bug",
			expected: []TokenWithText{},
		},
		{
			name: "priority then description",
			text: "(A) Fix a bug",
			expected: []TokenWithText{
				{Text: `(A)`, Role: todoTxtPriorityRole},
			},
		},
		{
			name: "completed task",
			text: "x Fix a bug",
			expected: []TokenWithText{
				{Text: `x Fix a bug`, Role: todoTxtCompletedTaskRole},
			},
		},
		{
			name:     "description starting with x",
			text:     "xFix",
			expected: []TokenWithText{},
		},
		{
			name: "single date",
			text: "2022-04-19 Fix a bug",
			expected: []TokenWithText{
				{Text: `2022-04-19`, Role: todoTxtDateRole},
			},
		},
		{
			name: "multiple dates",
			text: "2022-04-19 2022-01-02 Fix a bug",
			expected: []TokenWithText{
				{Text: `2022-04-19`, Role: todoTxtDateRole},
				{Text: `2022-01-02`, Role: todoTxtDateRole},
			},
		},
		{
			name: "project tag",
			text: "Fix a bug +project",
			expected: []TokenWithText{
				{Text: `+project`, Role: todoTxtProjectTagRole},
			},
		},
		{
			name: "context tag",
			text: "Fix a bug @context",
			expected: []TokenWithText{
				{Text: `@context`, Role: todoTxtContextTagRole},
			},
		},
		{
			name: "key-val tag",
			text: "Fix a bug status:review",
			expected: []TokenWithText{
				{Text: `status:`, Role: todoTxtKeyTagRole},
				{Text: `review`, Role: todoTxtValTagRole},
			},
		},
		{
			name:     "capital X at start is not completed",
			text:     "X Fix a bug",
			expected: []TokenWithText{},
		},
		{
			name: "full example, multiple tasks",
			text: `(A) Thank Mom for the meatballs @phone
(B) Schedule Goodwill pickup +GarageSale @phone
Post signs around the neighborhood +GarageSale
@GroceryStore Eskimo pies`,
			expected: []TokenWithText{
				{Text: `(A)`, Role: todoTxtPriorityRole},
				{Text: `@phone`, Role: todoTxtContextTagRole},
				{Text: `(B)`, Role: todoTxtPriorityRole},
				{Text: `+GarageSale`, Role: todoTxtProjectTagRole},
				{Text: `@phone`, Role: todoTxtContextTagRole},
				{Text: `+GarageSale`, Role: todoTxtProjectTagRole},
				{Text: `@GroceryStore`, Role: todoTxtContextTagRole},
			},
		},
		{
			name: "tasks without priorities",
			text: `Really gotta call Mom (A) @phone @someday
(b) Get back to the boss
(B)->Submit TPS report`,
			expected: []TokenWithText{
				{Text: `@phone`, Role: todoTxtContextTagRole},
				{Text: `@someday`, Role: todoTxtContextTagRole},
			},
		},
		{
			name:     "task without context",
			text:     "Email SoAndSo at soandso@example.com",
			expected: []TokenWithText{},
		},
		{
			name:     "task without project",
			text:     "Learn how to add 2+2",
			expected: []TokenWithText{},
		},
		{
			name: "tasks that aren't completed",
			text: `xylophone lesson
X 2012-01-01 Make resolutions
(A) x Find ticket prices`,
			expected: []TokenWithText{
				{Text: `2012-01-01`, Role: todoTxtDateRole},
				{Text: `(A)`, Role: todoTxtPriorityRole},
			},
		},
		{
			name: "priority must be at start of line",
			text: `
	(A) first
	(B) second
	(C) third
`,
			expected: []TokenWithText{},
		},
		{
			name: "indented with tags",
			text: `
	Write the design doc +myproject @laptop
`,
			expected: []TokenWithText{
				{Text: "+myproject", Role: todoTxtProjectTagRole},
				{Text: "@laptop", Role: todoTxtContextTagRole},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(TodoTxtParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}
