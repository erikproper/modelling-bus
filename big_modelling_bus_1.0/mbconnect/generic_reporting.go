/*
 *
 * Package: mbconnect
 * Layer:   generic
 * Module:  reporting
 *
 * ..... ... .. .
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), TU Wien, Austria
 *
 * Version of: XX.11.2025
 *
 */

package mbconnect

type (
	TErrorReporter func(string, error)
)
