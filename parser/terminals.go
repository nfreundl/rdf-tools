/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Edouard Martin Freundler
 */
package parser

// not a terminal

var CRorLF = newSet([]rune{0x0D, 0x0A}...)

// base terminals

var WS = newSet([]rune{0x20, 0x09, 0x0D, 0x0A}...)

var HEX = newSet().addRange(0, 9).addRange('a', 'f').addRange('A', 'F')

var PN_LOCAL_ESC = newSet([]rune{'_', '~', '.', '-', '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', '/', '?', '#', '@', '%'}...)

var PN_CHARS_BASE = func() RuneSet {
	ret := newSet()
	ret.addRange('A', 'Z').addRange('A', 'Z').addRange(0xC0, 0xD6).addRange(0xD8, 0xF6).addRange(0xF8, 0x02FF)
	ret.addRange(0x0370, 0x037D)
	ret.addRange(0x037F, 0x1FFF)
	ret.addRange(0x200C, 0x200D)
	ret.addRange(0x2070, 0x218F)
	ret.addRange(0x2C00, 0x2FEF)
	ret.addRange(0x3001, 0xD7FF)
	ret.addRange(0xF900, 0xFDCF)
	ret.addRange(0xFDF0, 0xFFFD)
	ret.addRange(0xFDF0, 0xFFFD)
	ret.addRange(0x00010000, 0x000EFFFF)
	return ret
}()

var PN_CHARS_U = PN_CHARS_BASE.copy().add('_')

var PN_CHARS = PN_CHARS_U.add('-').addRange('0', '9').add(0xB7).addRange(0x0300, 0x036F).addRange(0x203F, 0x2040)

const HEX2 = "[0-9]|[A-F]|[a-f]"

const UCHAR = "\\u(?:" + HEX2 + "){4}|\\U(?:" + HEX2 + "){4}"

const IRI2 = "<(?:[^\\x00-\\x20<>\"{}|^`\\] | (?:" + UCHAR + "))*>"

const ECHAR = `\[tbnrf\"']`

const WS2 = `[\x20\x09\x0d\xa]`
