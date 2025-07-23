package utils

import (
	"regexp"
	"strings"
	"testing"
	"time"

	vjailbreakv1alpha1 "github.com/platform9/vjailbreak/k8s/migration/api/v1alpha1"
	"github.com/platform9/vjailbreak/k8s/migration/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMigrationNameFromVMName(t *testing.T) {
	tests := []struct {
		name     string
		vmName   string
		expected string
	}{
		{
			name:     "Simple VM name",
			vmName:   "test-vm",
			expected: "migration-test-vm",
		},
		{
			name:     "Empty VM name",
			vmName:   "",
			expected: "migration-",
		},
		{
			name:     "Long VM name",
			vmName:   "this-is-a-very-long-virtual-machine-name-that-should-still-work",
			expected: "migration-this-is-a-very-long-virtual-machine-name-that-should-still-work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MigrationNameFromVMName(tt.vmName)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetMigrationConfigMapName(t *testing.T) {
	tests := []struct {
		name     string
		vmName   string
		expected string
	}{
		{
			name:     "Simple VM name",
			vmName:   "test-vm",
			expected: "migration-config-test-vm",
		},
		{
			name:     "Empty VM name",
			vmName:   "",
			expected: "migration-config-",
		},
		{
			name:     "Long VM name",
			vmName:   "this-is-a-very-long-virtual-machine-name-that-should-still-work",
			expected: "migration-config-this-is-a-very-long-virtual-machine-name-that-should-still-work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMigrationConfigMapName(tt.vmName)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestConvertToK8sName(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "Simple valid name",
			input:       "test-vm",
			expected:    "test-vm",
			expectError: false,
		},
		{
			name:        "Name with spaces",
			input:       "test vm with spaces",
			expected:    "test-vm-with-spaces",
			expectError: false,
		},
		{
			name:        "Name with underscores",
			input:       "test_vm_with_underscores",
			expected:    "test-vm-with-underscores",
			expectError: false,
		},
		{
			name:        "Mixed case name",
			input:       "TestVM",
			expected:    "testvm",
			expectError: false,
		},
		{
			name:        "Name with special characters",
			input:       "test-vm!@#$%^&*()",
			expected:    "test-vm",
			expectError: false,
		},
		{
			name:        "Very long name",
			input:       "this-is-a-very-long-virtual-machine-name-that-will-be-truncated-to-63-characters-which-is-the-k8s-limit",
			expected:    "this-is-a-very-long-virtual-machine-name-that-will-be-truncated",
			expectError: false,
		},
		{
			name:        "Name with trailing hyphen",
			input:       "test-vm-",
			expected:    "test-vm",
			expectError: false,
		},
		{
			name:        "Name with leading hyphen",
			input:       "-test-vm",
			expected:    "test-vm",
			expectError: false,
		},
		{
			name:        "Name with non-alphanumeric ending",
			input:       "test-vm-123-",
			expected:    "test-vm-123",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToK8sName(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result)
				}
				// Additional validation: ensure result meets k8s name requirements
				if len(result) > constants.K8sNameMaxLength {
					t.Errorf("Result length %d exceeds max length %d", len(result), constants.K8sNameMaxLength)
				}
				match, _ := regexp.MatchString("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$", result)
				if !match {
					t.Errorf("Result %q should match k8s name pattern", result)
				}
			}
		})
	}
}

func TestNewHostPathType(t *testing.T) {
	tests := []struct {
		name     string
		pathType string
		expected corev1.HostPathType
	}{
		{
			name:     "DirectoryOrCreate",
			pathType: string(corev1.HostPathDirectoryOrCreate),
			expected: corev1.HostPathDirectoryOrCreate,
		},
		{
			name:     "Directory",
			pathType: string(corev1.HostPathDirectory),
			expected: corev1.HostPathDirectory,
		},
		{
			name:     "FileOrCreate",
			pathType: string(corev1.HostPathFileOrCreate),
			expected: corev1.HostPathFileOrCreate,
		},
		{
			name:     "File",
			pathType: string(corev1.HostPathFile),
			expected: corev1.HostPathFile,
		},
		{
			name:     "Empty",
			pathType: "",
			expected: corev1.HostPathType(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewHostPathType(tt.pathType)
			if result == nil {
				t.Errorf("Expected non-nil result")
				return
			}
			if *result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, *result)
			}
		})
	}
}

func TestValidateMigrationPlan(t *testing.T) {
	now := metav1.Now()
	future := metav1.NewTime(now.Add(1 * time.Hour))
	past := metav1.NewTime(now.Add(-1 * time.Hour))

	tests := []struct {
		name          string
		migrationPlan *vjailbreakv1alpha1.MigrationPlan
		expectError   bool
		errorMessage  string
	}{
		{
			name: "Valid plan - cutover times valid",
			migrationPlan: &vjailbreakv1alpha1.MigrationPlan{
				Spec: vjailbreakv1alpha1.MigrationPlanSpec{
					MigrationPlanSpecPerVM: vjailbreakv1alpha1.MigrationPlanSpecPerVM{
						MigrationStrategy: vjailbreakv1alpha1.MigrationPlanStrategy{
							VMCutoverStart: now,
							VMCutoverEnd:   future,
						},
					},
					VirtualMachines: [][]string{{"vm1"}},
				},
			},
			expectError: false,
		},
		{
			name: "Invalid plan - cutover start after cutover end",
			migrationPlan: &vjailbreakv1alpha1.MigrationPlan{
				Spec: vjailbreakv1alpha1.MigrationPlanSpec{
					MigrationPlanSpecPerVM: vjailbreakv1alpha1.MigrationPlanSpecPerVM{
						MigrationStrategy: vjailbreakv1alpha1.MigrationPlanStrategy{
							VMCutoverStart: future,
							VMCutoverEnd:   past,
						},
					},
					VirtualMachines: [][]string{{"vm1"}},
				},
			},
			expectError:  true,
			errorMessage: "cutover start time is after cutover end time",
		},
		{
			name: "Invalid plan - advanced options with multiple VMs",
			migrationPlan: &vjailbreakv1alpha1.MigrationPlan{
				Spec: vjailbreakv1alpha1.MigrationPlanSpec{
					MigrationPlanSpecPerVM: vjailbreakv1alpha1.MigrationPlanSpecPerVM{
						MigrationStrategy: vjailbreakv1alpha1.MigrationPlanStrategy{
							VMCutoverStart: now,
							VMCutoverEnd:   future,
						},
						AdvancedOptions: vjailbreakv1alpha1.AdvancedOptions{
							GranularNetworks: []string{"custom-network"},
						},
					},
					VirtualMachines: [][]string{{"vm1", "vm2"}},
				},
			},
			expectError:  true,
			errorMessage: "advanced options can only be set for a single VM",
		},
		{
			name: "Valid plan - advanced options with single VM",
			migrationPlan: &vjailbreakv1alpha1.MigrationPlan{
				Spec: vjailbreakv1alpha1.MigrationPlanSpec{
					MigrationPlanSpecPerVM: vjailbreakv1alpha1.MigrationPlanSpecPerVM{
						MigrationStrategy: vjailbreakv1alpha1.MigrationPlanStrategy{
							VMCutoverStart: now,
							VMCutoverEnd:   future,
						},
						AdvancedOptions: vjailbreakv1alpha1.AdvancedOptions{
							GranularNetworks: []string{"custom-network"},
						},
					},
					VirtualMachines: [][]string{{"vm1"}},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMigrationPlan(tt.migrationPlan)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMessage != "" && !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("Error message %q does not contain expected message %q", err.Error(), tt.errorMessage)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestGenerateSha256Hash(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Empty string",
			input: "",
		},
		{
			name:  "Simple string",
			input: "test-vm",
		},
		{
			name:  "Complex string",
			input: "This is a complex string with spaces and special characters: !@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSha256Hash(tt.input)
			if result == "" {
				t.Errorf("Expected non-empty result")
			}
			if len(result) != 64 {
				t.Errorf("Expected length 64, got %d", len(result))
			}

			// Verify the hash is a valid hex string
			match, _ := regexp.MatchString("^[0-9a-f]+$", result)
			if !match {
				t.Errorf("Hash %q should be a valid hex string", result)
			}

			// Verify the last character is alphanumeric
			lastChar := result[len(result)-1]
			if !((lastChar >= '0' && lastChar <= '9') ||
				(lastChar >= 'a' && lastChar <= 'z') ||
				(lastChar >= 'A' && lastChar <= 'Z')) {
				t.Errorf("Last character %c should be alphanumeric", lastChar)
			}

			// Test idempotence
			result2 := GenerateSha256Hash(tt.input)
			if result != result2 {
				t.Errorf("Hash function should be idempotent: %q != %q", result, result2)
			}
		})
	}
}

func TestGetVMwareMachineNameForVMName(t *testing.T) {
	tests := []struct {
		name        string
		vmName      string
		expected    string
		expectError bool
	}{
		{
			name:        "Simple VM name",
			vmName:      "test-vm",
			expected:    "test-vm-" + GenerateSha256Hash("test-vm")[:constants.HashSuffixLength],
			expectError: false,
		},
		{
			name:        "Long VM name",
			vmName:      "this-is-a-very-long-virtual-machine-name-that-will-be-truncated",
			expectError: false,
		},
		{
			name:        "VM name with invalid characters",
			vmName:      "test_vm with special chars #$%",
			expected:    "test-vm-with-special-chars-" + GenerateSha256Hash("test-vm-with-special-chars")[:constants.HashSuffixLength],
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetVMwareMachineNameForVMName(tt.vmName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}

				if tt.expected != "" && result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result)
				}

				// Validate the result structure
				k8sName, err := ConvertToK8sName(tt.vmName)
				if err != nil {
					t.Errorf("Failed to convert VM name to k8s name: %v", err)
					return
				}

				nameWithoutHash := result[:len(result)-constants.HashSuffixLength-1] // Remove hash part
				expectedName := k8sName[:min(len(k8sName), constants.VMNameMaxLength)]
				if nameWithoutHash != expectedName {
					t.Errorf("Name part mismatch, expected %q, got %q", expectedName, nameWithoutHash)
				}

				// Validate the hash part
				expectedHash := GenerateSha256Hash(k8sName)[:constants.HashSuffixLength]
				resultHash := result[len(result)-constants.HashSuffixLength:]
				if resultHash != expectedHash {
					t.Errorf("Hash part mismatch, expected %q, got %q", expectedHash, resultHash)
				}

				// Ensure name is valid k8s name
				if len(result) > constants.K8sNameMaxLength {
					t.Errorf("Result length %d exceeds max length %d", len(result), constants.K8sNameMaxLength)
				}
			}
		})
	}
}

func TestGetJobNameForVMName(t *testing.T) {
	tests := []struct {
		name        string
		vmName      string
		expected    string
		expectError bool
	}{
		{
			name:        "Simple VM name",
			vmName:      "test-vm",
			expectError: false,
		},
		{
			name:        "Long VM name",
			vmName:      "this-is-a-very-long-virtual-machine-name-that-will-be-truncated",
			expectError: false,
		},
		{
			name:        "VM name with invalid characters",
			vmName:      "test_vm with special chars #$%",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetJobNameForVMName(tt.vmName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}

				// Validate result has the correct prefix
				if !strings.HasPrefix(result, "v2v-helper-") {
					t.Errorf("Result %q should have prefix 'v2v-helper-'", result)
				}

				// We don't need to verify vmware machine name here as we test it separately

				// For GetJobNameForVMName, we just validate that it follows the expected format
				// and is a valid K8s name without requiring the exact hash match
				if !strings.HasPrefix(result, "v2v-helper-") {
					t.Errorf("Result %q should start with v2v-helper-", result)
				}

				// Verify the name contains a hash suffix by checking for a hyphen followed by hash characters
				match, _ := regexp.MatchString("-[0-9a-f]+$", result)
				if !match {
					t.Errorf("Result %q should end with a hash pattern", result)
				}

				// Ensure name is valid k8s name and within length limits
				if len(result) > constants.K8sNameMaxLength {
					t.Errorf("Result length %d exceeds max length %d", len(result), constants.K8sNameMaxLength)
				}
			}
		})
	}
}
