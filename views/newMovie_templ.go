// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.894
package views

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	c "believer/movies/components"
	"believer/movies/types"
)

type NewMovieProps struct {
	ImdbID      string
	InWatchlist bool
	Movie       types.Movie
}

func NewMovie(props NewMovieProps) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Var2 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
			templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
			templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
			if !templ_7745c5c3_IsBuffer {
				defer func() {
					templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err == nil {
						templ_7745c5c3_Err = templ_7745c5c3_BufErr
					}
				}()
			}
			ctx = templ.InitializeContext(ctx)
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<form hx-post=\"/movie/new\" hx-indicator=\"#sending\" class=\"mx-auto flex max-w-xl flex-col gap-y-6 px-4 py-8\"><div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Var3 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
				templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
				templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
				if !templ_7745c5c3_IsBuffer {
					defer func() {
						templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
						if templ_7745c5c3_Err == nil {
							templ_7745c5c3_Err = templ_7745c5c3_BufErr
						}
					}()
				}
				ctx = templ.InitializeContext(ctx)
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "Back")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				return nil
			})
			templ_7745c5c3_Err = c.Link(c.LinkProps{Href: "/"}).Render(templ.WithChildren(ctx, templ_7745c5c3_Var3), templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "</div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if props.Movie.ID != 0 {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, "<div>Adding <strong>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var4 string
				templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(props.Movie.Title)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/newMovie.templ`, Line: 28, Col: 39}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "</strong></div>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			if props.ImdbID == "" {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 6, "<div class=\"flex flex-col gap-2 relative\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = c.Label("search", "Search").Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 7, "<input type=\"text\" hx-get=\"/movie/search\" hx-trigger=\"keyup changed delay:500ms\" hx-target=\"#search-results\" hx-validate=\"true\" minlength=\"3\" name=\"search\" id=\"search\" class=\"w-full rounded-sm border border-neutral-400 bg-transparent px-4 py-2 ring-offset-2 ring-offset-white focus:outline-hidden focus:ring-2 focus:ring-neutral-400 dark:border-neutral-700 dark:ring-offset-neutral-900 dark:focus:ring-neutral-500\"><div id=\"search-results\" class=\"text-xs empty:hidden rounded-sm p-2 outline-dashed outline-1 outline-neutral-400 dark:outline-neutral-500\"></div></div>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 8, "<div class=\"flex flex-col gap-2\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if props.ImdbID != "" {
				templ_7745c5c3_Err = c.Label("imdb_id", "IMDb ID").Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			} else {
				templ_7745c5c3_Err = c.Label("imdb_id", "IMDb ID or TMDB ID").Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 9, "<input required type=\"text\" hx-get=\"/movie/imdb\" hx-trigger=\"blur-sm changed\" hx-target=\"#movie-exists\" hx-validate=\"true\" name=\"imdb_id\" id=\"imdb_id\" class=\"w-full rounded-sm border border-neutral-400 bg-transparent px-4 py-2 ring-offset-2 ring-offset-white focus:outline-hidden focus:ring-2 focus:ring-neutral-400 dark:border-neutral-700 dark:ring-offset-neutral-900 dark:focus:ring-neutral-500\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if props.ImdbID != "" {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 10, " value=\"")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var5 string
				templ_7745c5c3_Var5, templ_7745c5c3_Err = templ.JoinStringErrs(props.ImdbID)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/newMovie.templ`, Line: 68, Col: 26}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var5))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 11, "\" readonly")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 12, "><div id=\"movie-exists\" class=\"text-xs empty:hidden lg:absolute lg:-right-52 lg:top-6 lg:w-48 lg:rounded-sm lg:p-2 lg:outline-dashed lg:outline-offset-4 lg:outline-neutral-500\"></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = c.Help("For example, https://www.imdb.com/title/tt0111161/, or just tt0111161.").Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 13, "</div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = c.NumberInput(c.NumberInputProps{Name: "rating", Label: "Rating", HelpText: "A value between 0 and 10", Min: 0, Max: 10, Required: true}).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if !props.InWatchlist {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 14, "<div class=\"flex gap-x-2 items-center\"><input type=\"checkbox\" name=\"watchlist\" id=\"watchlist\" class=\"rounded-sm accent-neutral-700 border border-neutral-700 bg-neutral-800 focus:outline-dashed focus:outline-offset-2 focus:outline-neutral-500\" _=\"on click if me.checked remove @required from #rating otherwise add @required='' to #rating\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = c.Label("watchlist", "Add to watchlist").Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 15, "</div>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 16, "<div class=\"flex flex-col gap-2 group\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = c.Label("review", "Review").Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 17, "<textarea name=\"review\" id=\"review\" class=\"w-full h-40 rounded-sm border border-neutral-400 bg-transparent px-4 py-2 ring-offset-2 ring-offset-white focus:outline-hidden focus:ring-2 focus:ring-neutral-400 dark:border-neutral-700 dark:ring-offset-neutral-900 dark:focus:ring-neutral-500 block\"></textarea><div class=\"flex gap-x-2 items-center\"><input type=\"checkbox\" name=\"review_private\" id=\"review_private\" class=\"rounded-sm accent-neutral-700 border border-neutral-700 bg-neutral-800 focus:outline-dashed focus:outline-offset-2 focus:outline-neutral-500\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = c.Label("review_private", "Review is private").Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 18, "</div></div><details><summary class=\"cursor-pointer\">Additional fields</summary><div class=\"mt-4 flex flex-col gap-y-6\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = c.DateTimeInput("watched_at", "Watched at", "Defaults to current time if left empty.", "").Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 19, "<div class=\"flex flex-col gap-2\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = c.Label("series", "Series").Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 20, "<input type=\"text\" name=\"series\" id=\"series\" list=\"series_list\" class=\"w-full rounded-sm border border-neutral-400 bg-transparent px-4 py-2 ring-offset-2 ring-offset-white focus:outline-hidden focus:ring-2 focus:ring-neutral-400 dark:border-neutral-700 dark:ring-offset-neutral-900 dark:focus:ring-neutral-500\" _=\"on keyup\n                if my.value is not empty\n                  add @required='' to #number_in_series\n                otherwise\n                  remove @required from #number_in_series\n              end\n              on change\n                if my.value is not empty\n                  put <datalist>option[value='${my.value}']/>'s @label into #series_name\n                  remove .hidden from #series_name \n                otherwise\n                  set #series_name's innerText to '' \n                  add .hidden to #series_name\n                \"><div class=\"text-xs text-content-secondary hidden\" id=\"series_name\"></div><div hx-get=\"/movie/new/series\" hx-swap=\"outerHTML\" hx-trigger=\"load\"></div></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = c.NumberInput(c.NumberInputProps{Name: "number_in_series", Label: "Number in series", Min: 0, Max: 1000}).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 21, "<div class=\"flex gap-x-2 items-center\"><input type=\"checkbox\" name=\"wilhelm_scream\" id=\"wilhelm_scream\" class=\"rounded-sm accent-neutral-700 border border-neutral-700 bg-neutral-800 focus:outline-dashed focus:outline-offset-2 focus:outline-neutral-500\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = c.Label("wilhelm_scream", "Wilhelm scream").Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 22, "</div></div></details><footer class=\"flex flex-col gap-y-4\"><div id=\"error\" class=\"empty:hidden text-rose-700 dark:text-rose-400 border border-dashed border-rose-700 dark:border-rose-400 p-4 rounded-sm\"></div><button class=\"rounded-sm bg-neutral-200 px-6 py-2 text-content-primary dark:bg-neutral-700\" type=\"submit\">Add</button><div id=\"sending\" class=\"htmx-indicator\">Sending...</div></footer></form>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return nil
		})
		templ_7745c5c3_Err = c.Layout(c.LayoutProps{Title: "Add movie"}).Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
