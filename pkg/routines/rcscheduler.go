package routines

import (
	"fmt"
    kclient "k8s.io/kubernetes/pkg/client/unversioned"
    klabels "k8s.io/kubernetes/pkg/labels"
    kfields "k8s.io/kubernetes/pkg/fields"

    kapi "k8s.io/kubernetes/pkg/api"
    kuapi "k8s.io/kubernetes/pkg/api/unversioned"
)

type RCScheduler struct{
    Client *kclient.Client
    Namespace string
    ReplicaCount int
    name string
    selector map[string]string
}

func (rcs *RCScheduler) Init()error{
    rcs.selector = map[string]string{
        "deployment": "khealth-rcscheduler",
    }
    rcs.name = "khealth-rcscheduler"

    rc := &kapi.ReplicationController{
        TypeMeta: kuapi.TypeMeta{
            Kind: "ReplicationController",
            APIVersion: "v1",
        },
        ObjectMeta: kapi.ObjectMeta{
            Name: rcs.name,
            Labels: rcs.selector,
        },
        Spec : kapi.ReplicationControllerSpec{
            Replicas : rcs.ReplicaCount,
            Selector: rcs.selector,
            Template: &kapi.PodTemplateSpec{
                ObjectMeta: kapi.ObjectMeta{
                    Labels: rcs.selector,
                },
                Spec: kapi.PodSpec{
                    RestartPolicy: kapi.RestartPolicyAlways,
                    Containers: []kapi.Container{
                        kapi.Container{
                            Name: "busybox-clock",
                            Image: "quay.io/aptible/busybox",
                            Command: []string{
                                "/bin/sh",
                                "-c",
                                "while true;do date;sleep 5;done",
                            },
                        },
                    },
                },
            },
        },
    }

    if _, err := rcs.Client.ReplicationControllers(rcs.Namespace).Create(rc); err != nil {
        return err
    }

    return nil
}

func (rcs *RCScheduler) Poll()error{
    rc, err := rcs.Client.ReplicationControllers(rcs.Namespace).Get(rcs.name)
    if err != nil {
        return err
    }

    if rc.Status.Replicas != rcs.ReplicaCount {
        return fmt.Errorf("Replica count mismatch: observed = %d, actual = %d",rc.Status.Replicas,rcs.ReplicaCount)
    }

    selector := klabels.SelectorFromSet(rcs.selector)
    if pods, err := rcs.Client.Pods(rcs.Namespace).List(selector,kfields.Everything()); err != nil {
        return err
    }else{
        for i := range(pods.Items){
            pod := pods.Items[i]
            if pod.Status.Phase != kapi.PodRunning {
                return fmt.Errorf("Pod %s = %s : %s : %s", pod.Name,pod.Status.Phase,pod.Status.Message, pod.Status.Reason)
            }
        }
    }

    return nil
}

func (rcs *RCScheduler) Cleanup()error{
    if err := rcs.Client.ReplicationControllers(rcs.Namespace).Delete(rcs.name); err != nil {
        return err
    }

    selector := klabels.SelectorFromSet(rcs.selector)

    if pods, err := rcs.Client.Pods(rcs.Namespace).List(selector,kfields.Everything()); err != nil {
        return fmt.Errorf("Error listing pods: %s",err)
    }else{
        for i := range(pods.Items){
            pod := pods.Items[i]
            if err := rcs.Client.Pods(rcs.Namespace).Delete(pod.Name,kapi.NewDeleteOptions(0)); err != nil  {
                return fmt.Errorf("Error deleting pod: %s",err)
            }
        }
    }
    return nil
}
