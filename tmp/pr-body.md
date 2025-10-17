# 🚨 Gosec Vulnerability Report for branch `main`
* File: /home/runner/work/vjailbreak/vjailbreak/v2v-cli/cmd/generate.go
  • Line: 75
  • Rule ID: G404
  • Details: Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand)
  • Confidence: MEDIUM
  • Severity: HIGH

* File: /home/runner/work/vjailbreak/vjailbreak/v2v-helper/vcenter/vcenterops.go
  • Line: 108-119
  • Rule ID: G402
  • Details: TLS MinVersion too low.
  • Confidence: HIGH
  • Severity: HIGH

* File: /home/runner/work/vjailbreak/vjailbreak/v2v-helper/migrate/migrate.go
  • Line: 944
  • Rule ID: G402
  • Details: TLS InsecureSkipVerify set true.
  • Confidence: HIGH
  • Severity: HIGH

* File: /home/runner/work/vjailbreak/vjailbreak/k8s/migration/pkg/utils/bmprovisionerutils.go
  • Line: 454
  • Rule ID: G402
  • Details: TLS InsecureSkipVerify set true.
  • Confidence: HIGH
  • Severity: HIGH

* File: /home/runner/work/vjailbreak/vjailbreak/k8s/migration/pkg/sdk/keystone/keystone.go
  • Line: 294
  • Rule ID: G402
  • Details: TLS InsecureSkipVerify set true.
  • Confidence: HIGH
  • Severity: HIGH

* File: /home/runner/work/vjailbreak/vjailbreak/k8s/migration/pkg/constants/constants.go
  • Line: 139
  • Rule ID: G101
  • Details: Potential hardcoded credentials
  • Confidence: LOW
  • Severity: HIGH

* File: /home/runner/work/vjailbreak/vjailbreak/k8s/migration/pkg/constants/constants.go
  • Line: 115
  • Rule ID: G101
  • Details: Potential hardcoded credentials
  • Confidence: LOW
  • Severity: HIGH

* File: /home/runner/work/vjailbreak/vjailbreak/k8s/migration/pkg/constants/constants.go
  • Line: 106
  • Rule ID: G101
  • Details: Potential hardcoded credentials
  • Confidence: LOW
  • Severity: HIGH

* File: /home/runner/work/vjailbreak/vjailbreak/k8s/migration/pkg/constants/constants.go
  • Line: 61
  • Rule ID: G101
  • Details: Potential hardcoded credentials
  • Confidence: LOW
  • Severity: HIGH

* File: /home/runner/work/vjailbreak/vjailbreak/k8s/migration/pkg/constants/constants.go
  • Line: 58
  • Rule ID: G101
  • Details: Potential hardcoded credentials
  • Confidence: LOW
  • Severity: HIGH

* File: /home/runner/work/vjailbreak/vjailbreak/k8s/migration/pkg/constants/constants.go
  • Line: 55
  • Rule ID: G101
  • Details: Potential hardcoded credentials
  • Confidence: LOW
  • Severity: HIGH

* File: /home/runner/work/vjailbreak/vjailbreak/k8s/migration/pkg/constants/constants.go
  • Line: 34
  • Rule ID: G101
  • Details: Potential hardcoded credentials
  • Confidence: LOW
  • Severity: HIGH

* File: /home/runner/work/vjailbreak/vjailbreak/k8s/migration/pkg/constants/constants.go
  • Line: 31
  • Rule ID: G101
  • Details: Potential hardcoded credentials
  • Confidence: LOW
  • Severity: HIGH

