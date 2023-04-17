/*
Copyright (C) 2021-2023, Kubefirst

This program is licensed under MIT.
See the LICENSE file for more details.
*/
package controller

import (
	"fmt"
	"strings"

	"github.com/kubefirst/runtime/pkg/k3d"
	"github.com/kubefirst/runtime/pkg/k8s"
	"github.com/kubefirst/runtime/pkg/terraform"
	log "github.com/sirupsen/logrus"
)

// RunUsersTerraform
func (clctrl *ClusterController) RunUsersTerraform() error {
	cl, err := clctrl.MdbCl.GetCluster(clctrl.ClusterName)
	if err != nil {
		return err
	}

	if !cl.UsersTerraformApplyCheck {
		kcfg := k8s.CreateKubeConfig(false, clctrl.ProviderConfig.Kubeconfig)
		// telemetryShim.Transmit(useTelemetryFlag, segmentClient, segment.MetricUsersTerraformApplyStarted, "")
		log.Info("applying users terraform")

		var vaultRootToken string
		secData, err := k8s.ReadSecretV2(kcfg.Clientset, "vault", "vault-unseal-secret")
		if err != nil {
			return err
		}

		vaultRootToken = secData["root-token"]

		tfEnvs := map[string]string{}
		tfEnvs["TF_VAR_email_address"] = "your@email.com"
		tfEnvs[fmt.Sprintf("TF_VAR_%s_token", clctrl.GitProvider)] = clctrl.GitToken
		tfEnvs["TF_VAR_vault_addr"] = k3d.VaultPortForwardURL
		tfEnvs["TF_VAR_vault_token"] = vaultRootToken
		tfEnvs["VAULT_ADDR"] = k3d.VaultPortForwardURL
		tfEnvs["VAULT_TOKEN"] = vaultRootToken
		tfEnvs[fmt.Sprintf("%s_TOKEN", strings.ToUpper(clctrl.GitProvider))] = clctrl.GitToken
		tfEnvs[fmt.Sprintf("%s_OWNER", strings.ToUpper(clctrl.GitProvider))] = clctrl.GitOwner

		tfEntrypoint := clctrl.ProviderConfig.GitopsDir + "/terraform/users"
		err = terraform.InitApplyAutoApprove(false, tfEntrypoint, tfEnvs)
		if err != nil {
			// telemetryShim.Transmit(useTelemetryFlag, segmentClient, segment.MetricUsersTerraformApplyStarted, err.Error())
			return err
		}
		log.Info("executed users terraform successfully")
		// telemetryShim.Transmit(useTelemetryFlag, segmentClient, segment.MetricUsersTerraformApplyCompleted, "")

		err = clctrl.MdbCl.UpdateCluster(clctrl.ClusterName, "users_terraform_apply_check", true)
		if err != nil {
			return err
		}
	}
	return nil
}
