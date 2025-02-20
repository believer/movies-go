// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.833
package views

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	c "believer/movies/components"
	"believer/movies/types"
)

func MoviesByYear(year string, movies types.Movies) templ.Component {
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
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<div>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var4 string
				templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(movies.NumberOfMovies())
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/moviesByYear.templ`, Line: 12, Col: 29}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "</div>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if len(movies) > 0 {
					templ_7745c5c3_Var5 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
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
						for _, movie := range movies {
							templ_7745c5c3_Var6 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
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
								templ_7745c5c3_Var7 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
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
									var templ_7745c5c3_Var8 string
									templ_7745c5c3_Var8, templ_7745c5c3_Err = templ.JoinStringErrs(movie.Title)
									if templ_7745c5c3_Err != nil {
										return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/moviesByYear.templ`, Line: 19, Col: 21}
									}
									_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var8))
									if templ_7745c5c3_Err != nil {
										return templ_7745c5c3_Err
									}
									return nil
								})
								templ_7745c5c3_Err = c.Link(c.LinkProps{Href: movie.LinkTo()}).Render(templ.WithChildren(ctx, templ_7745c5c3_Var7), templ_7745c5c3_Buffer)
								if templ_7745c5c3_Err != nil {
									return templ_7745c5c3_Err
								}
								templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, " ")
								if templ_7745c5c3_Err != nil {
									return templ_7745c5c3_Err
								}
								templ_7745c5c3_Err = c.Divider().Render(ctx, templ_7745c5c3_Buffer)
								if templ_7745c5c3_Err != nil {
									return templ_7745c5c3_Err
								}
								templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, " <span class=\"flex items-center gap-x-2 whitespace-nowrap text-sm tabular-nums relative top-1\">")
								if templ_7745c5c3_Err != nil {
									return templ_7745c5c3_Err
								}
								var templ_7745c5c3_Var9 string
								templ_7745c5c3_Var9, templ_7745c5c3_Err = templ.JoinStringErrs(movie.ISOReleaseDate())
								if templ_7745c5c3_Err != nil {
									return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/moviesByYear.templ`, Line: 23, Col: 32}
								}
								_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var9))
								if templ_7745c5c3_Err != nil {
									return templ_7745c5c3_Err
								}
								templ_7745c5c3_Err = Seen(SeenProps{
									Title:  "movie-by-year",
									Seen:   movie.Seen,
									ImdbId: movie.ImdbId,
									ID:     movie.ID,
								}).Render(ctx, templ_7745c5c3_Buffer)
								if templ_7745c5c3_Err != nil {
									return templ_7745c5c3_Err
								}
								templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "</span>")
								if templ_7745c5c3_Err != nil {
									return templ_7745c5c3_Err
								}
								return nil
							})
							templ_7745c5c3_Err = c.Li().Render(templ.WithChildren(ctx, templ_7745c5c3_Var6), templ_7745c5c3_Buffer)
							if templ_7745c5c3_Err != nil {
								return templ_7745c5c3_Err
							}
						}
						return nil
					})
					templ_7745c5c3_Err = c.Ol().Render(templ.WithChildren(ctx, templ_7745c5c3_Var5), templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				} else {
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 6, "<div>No movies this year</div>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				return nil
			})
			templ_7745c5c3_Err = StandardBody(year).Render(templ.WithChildren(ctx, templ_7745c5c3_Var3), templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return nil
		})
		templ_7745c5c3_Err = Layout(LayoutProps{Title: year}).Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
