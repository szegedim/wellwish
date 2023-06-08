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

var ActivationKey = "XPSZMNHVHDSOUOFNZBUQLBVVACMWASPLGXSQIZSDMXMDGJCKEXKCDQGLZWALMWWTJAFQILWYUMHSPZYSDHPDMSKVDXRR"
var ManagementKey = ""

var SiteName = "WellWish\nCloud Decision Engine"

// SiteUrl is the public facing external endpoint.
// This is typically a format like https://wellwish.example.com
var SiteUrl = ""

// Http11Port The container port that will face the public endpoint SiteUrl
var Http11Port = ":7777"

// NodePattern is easy to validate and a simple health script tells the nodes that are active.
// The system scans the cluster at startup.
// This is typically an internal node range
// You can use "10.55.0.0/21" for Kubernetes clusters
// GKE specific setting example:
// var NodePattern = "10.45.128.0/17"
// Standalone container
// var NodePattern = "127.0.0.1/32"
// Run Unit Tests
// var NodePattern = "http://127.0.0.1:77**"
// Suitable for local unit tests:
var NodePattern = "http://127.0.0.1:77**"

// StatefulBackupUrl is the standard backup location, if needed. Empty string, if it is not needed.
var StatefulBackupUrl = "http://127.0.0.1" + Http11Port

// DataRoot will normally be somewhere in /var/lib in the container to get backed up
var DataRoot = ""

var CompanyName = "Example Corporation (SAMPLE)"

var CompanyEmail = "hq@example.com"

var CompanyInfo = `Example Seller Inc.
1010 Corporate Avenue, San Francsico, CA, 55555, United States
TAX ID: 1234-56 Payment: ACH Routing# 12345 Account# 12345 https://example.com/12345
`
var CheckpointPeriod = 10 * 60 * time.Second
var PaymentPattern = "https://example.com/%s"

// UnitPrice per coin.
// We simulate something like a one dollar store.
// Everything uses the same coin as a vending machine.
// You can use services for different amounts.
// This helps to buy vouchers for multiple services in advance.
var UnitPrice = "USD 0.00"

// OrderPattern Please update the pattern based on your purchase order format.
var OrderPattern = `
Company: %s
Billing address: %s
Billing email: %s
We are placing an order for the following items.
We are ordering %s items of remoting vouchers for %s each.
The order total is %s.
The final amount includes Sales and Use Taxes of %s percent.
Net 30: Payment is due within 60 days of the invoice date.                 
Satisfaction guarantee. Cancel or refund within 30 days.
The vouchers are not for resale. They can only be used on this site.
Notes:
`

// InvoicePattern Please update the pattern with your locally regulated invoice format.
var InvoicePattern = `              INVOICE              

Pay To: %s
Date: %s        Invoice Number: %s

Payer: %s
Billing address: %s
Billing email: %s
Please pay the following remoting vouchers.
The payment term is Net 30.
Satisfaction guarantee.
Order is cancellable within 30 days.
Order is refundable within 30 days of order, if paid.

Ordered %s cloud vouchers for %s each.
               Invoiced Total %s.

The total amount includes sales and use taxes of 0 percent.
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

var RandomSalt = "XBGXTNTKIAVWBNHGODJGSSNUFBDIYPRYVKCFLYBFHPEWBRHQHYUWQLHHOPZLDZREJIAVPGEQMHOJFICSXNWADFHIHFRR"
var Simplify = false
