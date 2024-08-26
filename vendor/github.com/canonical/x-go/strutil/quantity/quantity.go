// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package quantity

import (
	"fmt"
	"math"

	"github.com/canonical/x-go/i18n"
)

// Common package strings with translation support
//
// Note: This may not be converted as part of global 'const' or 'var'
// variables, and also not inside 'init()'. The i18n implementation
// is only available after early initialisation.

func secs() string {
	return i18n.G("s")
}

func mins() string {
	return i18n.G("m")
}

func hours() string {
	return i18n.G("h")
}

func days() string {
	return i18n.G("d")
}

func years() string {
	return i18n.G("y")
}

// these are taken from github.com/chipaca/quantity with permission :-)

func FormatAmount(amount uint64, width int) string {
	if width < 0 {
		width = 5
	}
	max := uint64(5000)
	maxFloat := 999.5

	if width < 4 {
		width = 3
		max = 999
		maxFloat = 99.5
	}

	if amount <= max {
		pad := ""
		if width > 5 {
			pad = " "
		}
		return fmt.Sprintf("%*d%s", width-len(pad), amount, pad)
	}
	var prefix rune
	r := float64(amount)
	// zetta and yotta are me being pedantic: maxuint64 is ~18EB
	for _, prefix = range "kMGTPEZY" {
		r /= 1000
		if r < maxFloat {
			break
		}
	}

	width--
	digits := 3
	if r < 99.5 {
		digits--
		if r < 9.5 {
			digits--
			if r < .95 {
				digits--
			}
		}
	}
	precision := 0
	if (width - digits) > 1 {
		precision = width - digits - 1
	}

	s := fmt.Sprintf("%*.*f%c", width, precision, r, prefix)
	if r < .95 {
		return s[1:]
	}
	return s
}

func FormatBPS(n, sec float64, width int) string {
	if sec < 0 {
		sec = -sec
	}
	return FormatAmount(uint64(n/sec), width-2) + "B/s"
}

const (
	period = 365.25 // julian years (c.f. the actual orbital period, 365.256363004d)
)

func divmod(a, b float64) (q, r float64) {
	q = math.Floor(a / b)
	return q, a - q*b
}

// dt is seconds (as in the output of time.Now().Seconds())
func FormatDuration(dt float64) string {
	if dt < 60 {
		if dt >= 9.995 {
			return fmt.Sprintf("%.1f%.1s", dt, secs())
		} else if dt >= .9995 {
			return fmt.Sprintf("%.2f%.1s", dt, secs())
		}

		var prefix rune
		for _, prefix = range "mµn" {
			dt *= 1000
			if dt >= .9995 {
				break
			}
		}

		if dt > 9.5 {
			return fmt.Sprintf("%3.f%c%.1s", dt, prefix, secs())
		}

		return fmt.Sprintf("%.1f%c%.1s", dt, prefix, secs())
	}

	if dt < 600 {
		m, s := divmod(dt, 60)
		return fmt.Sprintf("%.f%.1s%02.f%.1s", m, mins(), s, secs())
	}

	dt /= 60 // dt now minutes

	if dt < 99.95 {
		return fmt.Sprintf("%3.1f%.1s", dt, mins())
	}

	if dt < 10*60 {
		h, m := divmod(dt, 60)
		return fmt.Sprintf("%.f%.1s%02.f%.1s", h, hours(), m, mins())
	}

	if dt < 24*60 {
		if h, m := divmod(dt, 60); m < 10 {
			return fmt.Sprintf("%.f%.1s%1.f%.1s", h, hours(), m, mins())
		}

		return fmt.Sprintf("%3.1f%.1s", dt/60, hours())
	}

	dt /= 60 // dt now hours

	if dt < 10*24 {
		d, h := divmod(dt, 24)
		return fmt.Sprintf("%.f%.1s%02.f%.1s", d, days(), h, hours())
	}

	if dt < 99.95*24 {
		if d, h := divmod(dt, 24); h < 10 {
			return fmt.Sprintf("%.f%.1s%.f%.1s", d, days(), h, hours())
		}
		return fmt.Sprintf("%4.1f%.1s", dt/24, days())
	}

	dt /= 24 // dt now days

	if dt < 2*period {
		return fmt.Sprintf("%4.0f%.1s", dt, days())
	}

	dt /= period // dt now years

	if dt < 9.995 {
		return fmt.Sprintf("%4.2f%.1s", dt, years())
	}

	if dt < 99.95 {
		return fmt.Sprintf("%4.1f%.1s", dt, years())
	}

	if dt < 999.5 {
		return fmt.Sprintf("%4.f%.1s", dt, years())
	}

	if dt > math.MaxUint64 || uint64(dt) == 0 {
		// TODO: figure out exactly what overflow causes the ==0
		return "ages!"
	}

	return FormatAmount(uint64(dt), 4) + years()
}
