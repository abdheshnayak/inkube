package fzf

import (
	//fzf "github.com/junegunn/fzf/src"
	"fmt"

	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/abdheshnayak/inkube/pkg/ui/text"
	"github.com/koki-develop/go-fzf"
	mfzf "github.com/koki-develop/go-fzf"
)

type Option mfzf.Option

func WithPrompt(prompt string) Option {
	return Option(mfzf.WithPrompt(fmt.Sprintf("%s %s ", prompt, text.Blue(":"))))
}

func FindOne[T any](items []T, itemFunc func(item T) string, options ...Option) (*T, error) {
	f, err := mfzf.New(func() []mfzf.Option {
		opts := make([]mfzf.Option, 0)
		for _, o := range options {
			opts = append(opts, mfzf.Option(o))
		}

		opts = append(opts, fzf.WithInputPlaceholder("search..."))
		return opts
	}()...)

	if err != nil {
		return nil, fn.NewE(err, "failed to create fzf")
	}

	idxs, err := f.Find(items, func(i int) string {
		return itemFunc(items[i])
	})

	if len(idxs) == 0 {
		return nil, fn.Error("you have not selected any item")
	}

	selectedIndex := idxs[0]

	return &items[selectedIndex], nil
}
