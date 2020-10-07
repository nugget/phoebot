load("@io_bazel_rules_k8s//k8s:object.bzl", "k8s_object")
load("@io_bazel_rules_k8s//k8s:objects.bzl", "k8s_objects")

DOCKER_REGISTRY = "index.docker.io"
DOCKER_ORG = "nugget"
NAME = "phoebot"
DOCKER_REPO = DOCKER_ORG + "/" + NAME

CLUSTER = "nuggethaus"
NAMESPACE = "phoebot"
USER = "nuggethaus"

def k8s_global():
    k8s_object(
        name = "namespace",
        kind = "namespace",
        template = ":k8s/namespace.yaml",
        cluster = CLUSTER,
        namespace = NAMESPACE,
        user = USER,
    )

def bot(botname):
    k8s_object(
        name = botname + "_deployment",
        kind = "deployment",
        template = ":k8s/deployment.yaml",
        cluster = CLUSTER,
        namespace = NAMESPACE,
        user = USER,
        images = {
            DOCKER_REGISTRY + "/" + DOCKER_REPO: ":image",
        },
        substitutions = {
            "{BOTNAME}": botname,
        }
    )

    k8s_object(
        name = botname +  "_configmap",
        kind = "configmap",
        template = ":k8s/" + botname + "_configmap.yaml",
        cluster = CLUSTER,
        namespace = NAMESPACE,
        user = USER,
    )

    k8s_objects(
        name = botname,
        objects = [
            ":namespace",
            ":" + botname + "_configmap",
            ":" + botname + "_deployment",
        ]
    )
