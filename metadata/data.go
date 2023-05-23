package metadata

import "time"

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// You need to update this file to fine tune your implementation of WellWish
// The design goal is to omit any json, yaml, xml configuration files being (i.e. devops)
// The solution runs as go run anyway making .go and .yaml obsolete
// Traditionally these were separate having different teams to fine tune and slow compilation
// Nowadays Golang building is fast, and it is more secure to have the open source code shipped in the container
// This makes it easy for distributors to verify what is actually running

// This is an example to fine tune a new docker image
// (docker pull registry.gitlab.com/eper.io/<project> | grep 'Downloaded newer image') && docker build -t example.com/wellwish . && docker push example.com/wellwish

var ActivationKey = "HQVNXPNHYXZISISZALKCPXKGXYOJUKZFHQXQJWEYGYDZCWWMXSDSZFQEDAMZGWLOESBVDKFWBAHHOVJGYJCXOEHXVFFR"
var ManagementKey = ""

var SiteName = "WellWish\nCloud Decision Engine"

var SiteUrl = "http://127.0.0.1:7777"

// NodePattern is easy to validate and a simple health script tells the nodes that are active.
// The system scans the cluster at startup.
// This is typically an internal node range
var NodePattern = "http://127.0.0.1:77**"

// StatefulBackupUrl is the standard backup location, if needed. Empty string, if it is not needed.
var StatefulBackupUrl = "http://127.0.0.1:7777"

var Http11Port = ":7777"

// DataRoot will normally be somewhere in /var/lib in the container to get backed up
var DataRoot = ""

var CompanyName = "Example Corporation (SAMPLE)"

var CompanyEmail = "hq@example.com"

var CompanyInfo = `Example Inc.
1010 Corporate Avenue, San Jose, CA, 55555, USA
TAX ID: 1234-56 Payment: ACH Routing# 12345 Account# 12345 https://example.com/12345
`
var CheckpointPeriod = 10 * 60 * time.Second
var PaymentPattern = "https://example.com/%s"
var UnitPrice = "USD 1.03"

// Please update the pattern based on your purchase order format.
var OrderPattern = `
Company: %s
Billing address: %s
Billing email: %s
Our company places the following order.
The payment term is Net 30.                 
Ordering %s remoting vouchers for %s each.

The order total is %s.
The final amount includes Sales Tax of %s percent.
Satisfaction guarantee. Cancel or refund within 30 days.
Notes:
`

// InvoicePattern Please update the pattern with your locally regulated invoice format.
var InvoicePattern = `              INVOICE              

Payee: %s
Date: %s        Invoice Number: %s

Payer: %s
Billing address: %s
Billing email: %s
Please pay the following remoting vouchers.
The payment term is Net 30.
Satisfaction guarantee.
Order is cancellable within 30 days.
Order is refundable within 30 days, if paid.

Ordered %s cloud vouchers for %s each
               Invoiced Total %s.

The total amount includes sales tax of 0 percent.
Order Status:
%s.
`

var VoucherPattern = `              SERVICE VOUCHER              

From: %s
Issue Date: %s
This voucher can be used at the servicing company listed above.
It is valid for 365 days from the time of issuance.
Invoice: %s
The voucher status is %s.
`

var RandomSalt = "AAKNVZJKRIOWSVLFASSGQXWOTDXVAKMGTJOKTQBAKLFCOKMEIDQRSKCTQLTDOVHXZLUKZKALBFDIXBSZHQCRGWZFYILR"
