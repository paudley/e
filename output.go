/* Copyright (C) 2022, 2023, 2024 by Blackcat Informatics® Inc.
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

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
		errorpath := ErrorPathJSON{
			Caller: fmt.Sprintf("%s:%d/%s -> %s", p.FileName, p.LineNumber, p.FuncName, p.Msg),
		}

		vals := p.Values()

		for _, v := range vals {
			errorpath.Values = append(errorpath.Values, c.Sdump(v))
		}

		eps = append(eps, errorpath)
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

	for i, pathe := range path {
		col := c.Orange
		if i == (len(path) - 1) {
			col = c.Yellow
		}

		sum += fmt.Sprintf("%s %s:%s/%s -> %s\n",
			c.Red.Sprint("- ->"),
			col.Sprint(pathe.FileName),
			col.Sprintf("%d", pathe.LineNumber),
			col.Sprint(pathe.FuncName),
			c.White.Sprint(pathe.Msg),
		)

		vals := pathe.Values()
		for _, val := range vals {
			switch valV := val.(type) {
			case V:
				switch valV.K {
				case "db_error":
					sum += fmt.Sprintf("%s %s %s: %s\n",
						c.Red.Sprint("-"),
						c.Magenta.Sprint("--$"),
						c.WhiteOnCyan.Sprint(" DB Error "),
						c.White.Sprint(valV.I))
				case "io_error":
					sum += fmt.Sprintf("%s %s %s: %s\n",
						c.Red.Sprint("-"),
						c.Magenta.Sprint("--$"),
						c.WhiteOnRed.Sprint(" IO Error "),
						c.White.Sprint(valV.I))
				case "json":
					str, convOK := valV.I.(string)
					if convOK {
						sum += fmt.Sprintf("%s %s %s: %s\n",
							c.Red.Sprint("-"),
							c.Magenta.Sprint("--$"),
							c.WhiteOnGreen.Sprint(" json "),
							c.SimpleColorString("json", str))
					}
				case "sql":
					str, convOK := valV.I.(string)
					if convOK {
						sum += fmt.Sprintf("%s %s %s: %s\n",
							c.Red.Sprint("-"),
							c.Magenta.Sprint("--$"),
							c.WhiteOnBlue.Sprint(" sql "),
							c.SimpleColorString("sql", str))
					}
				case "validation":
					sum += fmt.Sprintf("%s %s %s: %s\n",
						c.Red.Sprint("-"),
						c.Magenta.Sprint("--$"),
						c.BlackOnYellow.Sprint(" validation "),
						c.White.Sprint(valV.I))
				default:
					sum += fmt.Sprintf("%s %s %s => %s",
						c.Red.Sprint("-"),
						c.Magenta.Sprint("--$"),
						c.WhiteOnMagenta.Sprintf(" %s ", valV.K),
						c.SdumpColorSimple(valV.I))
				}
			default:
				sum += fmt.Sprintf("%s %s %s",
					c.Red.Sprint("-"),
					c.Magenta.Sprint("--$"),
					c.SdumpColorSimple(val))
			}
		}
	}

	sum += c.Red.Sprint("!! -----------------------------Error-- !!\n")

	return sum
}
