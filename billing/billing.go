package billing

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Billing is a nice pay later feature.
// Consult with a local professional for licensing.
// Customers can use an office cluster right away by typing billing details.
// They have the option to try features this way, and they can pay later.
// They can also cancel the order or get a refund, if they are not satisfied.
// The checkout feature allows to fill in the details and the number of vouchers needed.
// The invoice feature allows accepting or rejecting an invoice.
// The invoice URL can be sent in emails.
// The invoice can be used to download vouchers.
// Vouchers are not securities as they cannot be resold,
// and they are tied to the company that ordered.
// They are simply a proof of payment.
// Vouchers are downloaded for an invoice into a coin file.
// You can store the coin file on an usb stick to be more secure.
// We do not use any browser features.
// Tpm and extra encryption helps to deter conservative hackers.
// The reality is that if a malware scrapes the memory it can still grab these.
// We opt for storing them on usb stick.
// The coin file is a list of valid and used vouchers.
// Using the coin file means uploading it on one of the pages or sending it over an api.
// The voucher key is unique, and it does not have any hints where to use.
// We believe this makes it more secure.
// Traditional tokens have private keys and a base64 encoded json structure.
// However, this helps attackers to figure out where to use the token,
// if it was collected from the internet.
// If we use random numbers, just the number requires sending it to servers to check.
// This reveals the number deterring any for profit attackers, also
// it can reveal the random theft for security professionals.
// We rather use just the URL to identify our site, so that the random is not sent to other sites.

func Setup() {
	setupVoucher()
	setupCheckout()
	setupInvoice()
}
