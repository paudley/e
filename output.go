// Copyright (C) 2022, 2023, 2024 by Blackcat Informatics® Inc.
//
// nolint: varnamelen,funlen,cyclop
package e

import (
	"fmt"

	c "github.com/paudley/colorout"
)

type ErrorPathJSON struct {
	Caller string
	Values []string
}

type ErrorJSON struct {
	Context string
	Message string
	Path    []ErrorPathJSON
}

func (e Error) JSON() map[string]any {
	ret := make(map[string]any)
	ret["Kind"] = "errorBacktrace"
	ret["Context"] = e.OriginContextString()
	ret["Message"] = e.LastMessage()

	eps := []ErrorPathJSON{}
	path := e.Path()

	for _, p := range path {
		ep := ErrorPathJSON{
			Caller: fmt.Sprintf("%s:%d/%s -> %s", p.FileName, p.LineNumber, p.FuncName, p.Msg),
		}

		vals := p.Values()

		for _, v := range vals {
			ep.Values = append(ep.Values, c.Sdump(v))
		}

		eps = append(eps, ep)
	}

	ret["Path"] = eps

	return ret
}

// SummarizeConsole prepares a console friendly version of the error suitable for
// printing.
func (e Error) SummarizeConsole() string {
	msg := e.LastMessage()
	path := e.Path()
	sum := fmt.Sprintf("%s %s\n",
		c.Red.Sprint(`!! --Error--------------------------- !!
- err:`),
		c.White.Sprint(msg))

	octx := e.OriginContextString()
	if octx != "" {
		sum += fmt.Sprintf("%s %s\n", c.Red.Sprint("- ->"), c.Green.Sprint(octx))
	}

	for i, p := range path {
		col := c.Orange
		if i == (len(path) - 1) {
			col = c.Yellow
		}

		sum += fmt.Sprintf("%s %s:%s/%s -> %s\n",
			c.Red.Sprint("- ->"),
			col.Sprint(p.FileName),
			col.Sprintf("%d", p.LineNumber),
			col.Sprint(p.FuncName),
			c.White.Sprint(p.Msg),
		)

		vals := p.Values()
		for _, v := range vals {
			switch vV := v.(type) {
			case V:
				switch vV.K {
				case "db_error":
					sum += fmt.Sprintf("%s %s %s: %s\n",
						c.Red.Sprint("-"),
						c.Magenta.Sprint("--$"),
						c.WhiteOnCyan.Sprint(" DB Error "),
						c.White.Sprint(vV.I))
				case "io_error":
					sum += fmt.Sprintf("%s %s %s: %s\n",
						c.Red.Sprint("-"),
						c.Magenta.Sprint("--$"),
						c.WhiteOnRed.Sprint(" IO Error "),
						c.White.Sprint(vV.I))
				case "json":
					s, convOK := vV.I.(string)
					if convOK {
						sum += fmt.Sprintf("%s %s %s: %s\n",
							c.Red.Sprint("-"),
							c.Magenta.Sprint("--$"),
							c.WhiteOnGreen.Sprint(" json "),
							c.SimpleColorString("json", s))
					}
				case "sql":
					s, convOK := vV.I.(string)
					if convOK {
						sum += fmt.Sprintf("%s %s %s: %s\n",
							c.Red.Sprint("-"),
							c.Magenta.Sprint("--$"),
							c.WhiteOnBlue.Sprint(" sql "),
							c.SimpleColorString("sql", s))
					}
				case "validation":
					sum += fmt.Sprintf("%s %s %s: %s\n",
						c.Red.Sprint("-"),
						c.Magenta.Sprint("--$"),
						c.BlackOnYellow.Sprint(" validation "),
						c.White.Sprint(vV.I))
				default:
					sum += fmt.Sprintf("%s %s %s => %s",
						c.Red.Sprint("-"),
						c.Magenta.Sprint("--$"),
						c.WhiteOnMagenta.Sprintf(" %s ", vV.K),
						c.SdumpColorSimple(vV.I))
				}
			default:
				sum += fmt.Sprintf("%s %s %s",
					c.Red.Sprint("-"),
					c.Magenta.Sprint("--$"),
					c.SdumpColorSimple(v))
			}
		}
	}

	sum += c.Red.Sprint("!! -----------------------------Error-- !!\n")

	return sum
}
