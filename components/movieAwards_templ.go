// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.894
package components

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"believer/movies/types"
	"believer/movies/utils"
	"fmt"
)

type SectionProps interface {
	Href() templ.SafeURL
	NumberOfAwards() int
	Subtitle() string
	Title() string
	Wins() int
}

func awardSection(props SectionProps) templ.Component {
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
		if props.NumberOfAwards() > 0 {
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
				templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				return nil
			})
			templ_7745c5c3_Err = SectionNew(SectionTitleProps{
				Href:     props.Href(),
				Title:    props.Title(),
				Subtitle: props.Subtitle(),
			}).Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		return nil
	})
}

type MovieAwardsProps struct {
	Awards []types.Award
	Year   string
	Won    int
}

func (p MovieAwardsProps) Subtitle() string {
	wins := p.Wins()

	if p.Wins() == 0 {
		return fmt.Sprintf("%s", utils.PluralMessage(utils.NominationKey, p.NumberOfAwards()))
	}

	return fmt.Sprintf("%s / %s", utils.PluralMessage(utils.NominationKey, p.NumberOfAwards()), utils.PluralMessage(utils.WinKey, wins))
}

func (p MovieAwardsProps) NumberOfAwards() int {
	return len(p.Awards)
}

func (p MovieAwardsProps) Wins() int {
	return p.Won
}

func (p MovieAwardsProps) Title() string {
	return fmt.Sprintf("Academy Awards %s", p.Year)
}

func (p MovieAwardsProps) Href() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf("/awards/year/%s", p.Year))
}

func (p MovieAwardsProps) NominationMsg() string {
	return utils.PluralMessage(utils.NominationKey, p.NumberOfAwards())
}

func (p MovieAwardsProps) WinMsg() string {
	return utils.PluralMessage(utils.WinKey, p.Wins())
}

func MovieAwards(props MovieAwardsProps) templ.Component {
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
		templ_7745c5c3_Var3 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var3 == nil {
			templ_7745c5c3_Var3 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Var4 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
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
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<ul class=\"flex flex-col gap-2\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			for _, award := range props.Awards {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "<li class=\"col-span-2 flex items-baseline justify-between gap-x-4\"><div>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var5 string
				templ_7745c5c3_Var5, templ_7745c5c3_Err = templ.JoinStringErrs(award.Category)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `components/movieAwards.templ`, Line: 75, Col: 22}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var5))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var6 string
				templ_7745c5c3_Var6, templ_7745c5c3_Err = templ.JoinStringErrs(" ")
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `components/movieAwards.templ`, Line: 75, Col: 29}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var6))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "<div class=\"inline-flex\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if len(award.Nominees) > 0 {
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, "<span>(</span> ")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					for i, n := range award.Nominees {
						if n.ID != 0 {
							templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "<a class=\"border-b border-dashed border-neutral-500 focus-visible:outline-1 focus-visible:rounded-xs focus-visible:outline-dashed focus-visible:outline-offset-2 focus-visible:outline-neutral-400 dark:border-neutral-400 dark:focus-visible:outline-neutral-600 whitespace-nowrap\" href=\"")
							if templ_7745c5c3_Err != nil {
								return templ_7745c5c3_Err
							}
							var templ_7745c5c3_Var7 templ.SafeURL = n.LinkTo()
							_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(string(templ_7745c5c3_Var7)))
							if templ_7745c5c3_Err != nil {
								return templ_7745c5c3_Err
							}
							templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 6, "\">")
							if templ_7745c5c3_Err != nil {
								return templ_7745c5c3_Err
							}
							var templ_7745c5c3_Var8 string
							templ_7745c5c3_Var8, templ_7745c5c3_Err = templ.JoinStringErrs(n.Name)
							if templ_7745c5c3_Err != nil {
								return templ.Error{Err: templ_7745c5c3_Err, FileName: `components/movieAwards.templ`, Line: 87, Col: 19}
							}
							_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var8))
							if templ_7745c5c3_Err != nil {
								return templ_7745c5c3_Err
							}
							templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 7, "</a>")
							if templ_7745c5c3_Err != nil {
								return templ_7745c5c3_Err
							}
						} else {
							var templ_7745c5c3_Var9 string
							templ_7745c5c3_Var9, templ_7745c5c3_Err = templ.JoinStringErrs(n.Name)
							if templ_7745c5c3_Err != nil {
								return templ.Error{Err: templ_7745c5c3_Err, FileName: `components/movieAwards.templ`, Line: 90, Col: 18}
							}
							_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var9))
							if templ_7745c5c3_Err != nil {
								return templ_7745c5c3_Err
							}
						}
						templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 8, " ")
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
						if i < len(award.Nominees)-1 {
							templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 9, "<span class=\"mr-1\">")
							if templ_7745c5c3_Err != nil {
								return templ_7745c5c3_Err
							}
							var templ_7745c5c3_Var10 string
							templ_7745c5c3_Var10, templ_7745c5c3_Err = templ.JoinStringErrs(", ")
							if templ_7745c5c3_Err != nil {
								return templ.Error{Err: templ_7745c5c3_Err, FileName: `components/movieAwards.templ`, Line: 94, Col: 17}
							}
							_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var10))
							if templ_7745c5c3_Err != nil {
								return templ_7745c5c3_Err
							}
							templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 10, "</span>")
							if templ_7745c5c3_Err != nil {
								return templ_7745c5c3_Err
							}
						}
					}
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 11, " <span>)</span> ")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				if award.Detail.Valid {
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 12, "(")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var11 string
					templ_7745c5c3_Var11, templ_7745c5c3_Err = templ.JoinStringErrs(award.Detail.String)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `components/movieAwards.templ`, Line: 103, Col: 30}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var11))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 13, ")")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 14, "</div></div>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = Divider().Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 15, "<span>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if award.Winner {
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 16, "Won")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				} else {
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 17, "Nominated")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 18, "</span></li>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 19, "</ul>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = Divider().Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 20, " <div class=\"flex flex-col gap-y-1 text-xs\"><div class=\"flex gap-x-1\"><span>All movies with </span> <a class=\"border-b border-dashed border-neutral-500 focus-visible:outline-1 focus-visible:rounded-xs focus-visible:outline-dashed focus-visible:outline-offset-2 focus-visible:outline-neutral-400 dark:border-neutral-400 dark:focus-visible:outline-neutral-600 whitespace-nowrap\" href=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var12 templ.SafeURL = templ.SafeURL(fmt.Sprintf("/awards/%d?nominations=true", props.NumberOfAwards()))
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(string(templ_7745c5c3_Var12)))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 21, "\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var13 string
			templ_7745c5c3_Var13, templ_7745c5c3_Err = templ.JoinStringErrs(props.NominationMsg())
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `components/movieAwards.templ`, Line: 127, Col: 28}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var13))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 22, "</a> ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if props.Won > 0 {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 23, "<span>or</span><div><a class=\"border-b border-dashed border-neutral-500 focus-visible:outline-1 focus-visible:rounded-xs focus-visible:outline-dashed focus-visible:outline-offset-2 focus-visible:outline-neutral-400 dark:border-neutral-400 dark:focus-visible:outline-neutral-600 whitespace-nowrap\" href=\"")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var14 templ.SafeURL = templ.SafeURL(fmt.Sprintf("/awards/%d", props.Won))
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(string(templ_7745c5c3_Var14)))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 24, "\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var15 string
				templ_7745c5c3_Var15, templ_7745c5c3_Err = templ.JoinStringErrs(props.WinMsg())
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `components/movieAwards.templ`, Line: 134, Col: 23}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var15))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 25, "</a></div>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 26, "</div></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return nil
		})
		templ_7745c5c3_Err = awardSection(props).Render(templ.WithChildren(ctx, templ_7745c5c3_Var4), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

type PersonAwardsProps struct {
	Awards map[string][]types.Award
	Won    int
}

func (p PersonAwardsProps) Subtitle() string {
	return fmt.Sprintf("%s / %s", utils.PluralMessage(utils.NominationKey, p.NumberOfAwards()), utils.PluralMessage(utils.WinKey, p.Wins()))
}

func (p PersonAwardsProps) NumberOfAwards() int {
	awards := 0

	for _, c := range p.Awards {
		awards = awards + len(c)
	}

	return awards
}

func (p PersonAwardsProps) Wins() int {
	return p.Won
}

func (p PersonAwardsProps) Title() string {
	return "Academy Awards"
}

func (p PersonAwardsProps) Href() templ.SafeURL {
	return templ.SafeURL("")
}

func PersonAwards(props PersonAwardsProps) templ.Component {
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
		templ_7745c5c3_Var16 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var16 == nil {
			templ_7745c5c3_Var16 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Var17 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
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
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 27, "<section class=\"flex flex-col gap-y-6\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			for category, awards := range props.Awards {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 28, "<section class=\"flex flex-col gap-y-4\"><h3 class=\"font-bold text-xs\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var18 string
				templ_7745c5c3_Var18, templ_7745c5c3_Err = templ.JoinStringErrs(category)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `components/movieAwards.templ`, Line: 178, Col: 45}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var18))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 29, "</h3><ul class=\"flex flex-col gap-y-2\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				for _, award := range awards {
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 30, "<li class=\"col-span-2 flex items-end justify-between gap-x-4\"><span class=\"flex gap-x-2\">")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					if award.Title.Valid {
						templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 31, "<a class=\"border-b border-dashed border-neutral-500 focus-visible:outline-1 focus-visible:rounded-xs focus-visible:outline-dashed focus-visible:outline-offset-2 focus-visible:outline-neutral-400 dark:border-neutral-400 dark:focus-visible:outline-neutral-600 whitespace-nowrap\" href=\"")
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
						var templ_7745c5c3_Var19 templ.SafeURL = award.LinkToMovie()
						_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(string(templ_7745c5c3_Var19)))
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
						templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 32, "\">")
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
						var templ_7745c5c3_Var20 string
						templ_7745c5c3_Var20, templ_7745c5c3_Err = templ.JoinStringErrs(award.Title.String)
						if templ_7745c5c3_Err != nil {
							return templ.Error{Err: templ_7745c5c3_Err, FileName: `components/movieAwards.templ`, Line: 188, Col: 31}
						}
						_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var20))
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
						templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 33, "</a> ")
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
					}
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 34, "<a class=\"border-b border-dashed border-neutral-500 focus-visible:outline-1 focus-visible:rounded-xs focus-visible:outline-dashed focus-visible:outline-offset-2 focus-visible:outline-neutral-400 dark:border-neutral-400 dark:focus-visible:outline-neutral-600 whitespace-nowrap\" href=\"")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var21 templ.SafeURL = award.LinkToYear()
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(string(templ_7745c5c3_Var21)))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 35, "\">(")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var22 string
					templ_7745c5c3_Var22, templ_7745c5c3_Err = templ.JoinStringErrs(award.Year)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `components/movieAwards.templ`, Line: 195, Col: 23}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var22))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 36, ")</a> ")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					if award.Detail.Valid {
						templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 37, "(")
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
						var templ_7745c5c3_Var23 string
						templ_7745c5c3_Var23, templ_7745c5c3_Err = templ.JoinStringErrs(award.Detail.String)
						if templ_7745c5c3_Err != nil {
							return templ.Error{Err: templ_7745c5c3_Err, FileName: `components/movieAwards.templ`, Line: 198, Col: 32}
						}
						_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var23))
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
						templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 38, ")")
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
					}
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 39, "</span>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					templ_7745c5c3_Err = Divider().Render(ctx, templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 40, "<span>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					if award.Winner {
						templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 41, "Won")
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
					} else {
						templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 42, "Nominated")
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
					}
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 43, "</span></li>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 44, "</ul></section>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 45, "</section>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return nil
		})
		templ_7745c5c3_Err = awardSection(props).Render(templ.WithChildren(ctx, templ_7745c5c3_Var17), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
