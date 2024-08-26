// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2022 Canonical Ltd
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

// Package i18n provides an implementation agnostic API to be used in
// both library packages and application packages importing this library.
// The application has the option to:
//
// 1. Initialise this package by implementing the i18n interface
// 2. Leave the package uninitialised, disabling translation
//
// Warning:
//
// None of the i18n functionality may be used during early
// initialisation code such as when defining a package 'const', 'var'
// or 'init()'. This will prevent the application initialising
// the i18n interface before it gets used in packages, and will
// result in translation being disabled until the initialisation
// setup is complete.

package i18n

// The following public i18n function variables define the
// internationalisation marker API available. An implementation specific
// version can be provided by the application or alternatively if left
// unmodified the default functions will by used (translation disabled).
var (
	G  = GDefault
	NG = NGDefault
)

// GDefault is the fallback implementation. This will simply return
// the provided string untranslated.
func GDefault(msgid string) string {
	return msgid
}

// NGDefault is the fallback implementation. This will simply return
// the provided singular or plural string (depending on 'n')
// untranslated.
func NGDefault(msgid string, msgidPlural string, n int) string {
	if n == 1 {
		// Singular
		return msgid
	}

	// Plural
	return msgidPlural
}
