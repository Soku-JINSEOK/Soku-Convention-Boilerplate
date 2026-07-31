import copy
import importlib.util
import pathlib
import unittest

SCRIPT = pathlib.Path(__file__).with_name("verify-cloud-build-logging-plan.py")
SPEC = importlib.util.spec_from_file_location("plan_verifier", SCRIPT)
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


def valid_plan():
    resources = []
    after_values = {
        "google_logging_project_bucket_config.cloud_build_validation": {
            "location": "asia-northeast1",
            "retention_days": 30,
        },
        "google_logging_project_sink.cloud_build_validation": {
            "unique_writer_identity": False,
        },
        "google_logging_project_exclusion.default_disabled": {
            "name": "_Default",
            "disabled": True,
        },
    }
    for address, expected in MODULE.EXPECTED.items():
        resources.append(
            {
                "address": address,
                "type": expected["type"],
                "change": {"actions": ["create"], "after": after_values[address]},
            }
        )
    return {"resource_changes": resources}


class PlanVerifierTest(unittest.TestCase):
    def test_accepts_exact_three_resource_create(self):
        self.assertEqual(MODULE.verify_plan(valid_plan()), [])

    def test_rejects_large_or_unexpected_plan(self):
        plan = valid_plan()
        for index in range(39):
            plan["resource_changes"].append(
                {
                    "address": f"google_storage_bucket.unexpected_{index}",
                    "type": "google_storage_bucket",
                    "change": {"actions": ["create"], "after": {}},
                }
            )
        self.assertTrue(MODULE.verify_plan(plan))

    def test_rejects_iam_update_delete_required_and_enabled_exclusion(self):
        mutations = []
        plan = valid_plan()
        plan["resource_changes"][0]["change"]["actions"] = ["update"]
        mutations.append(plan)
        plan = valid_plan()
        plan["resource_changes"][1]["change"]["actions"] = ["delete"]
        mutations.append(plan)
        plan = valid_plan()
        plan["resource_changes"].append(
            {
                "address": "google_project_iam_member.writer",
                "type": "google_project_iam_member",
                "change": {"actions": ["create"], "after": {}},
            }
        )
        mutations.append(plan)
        plan = valid_plan()
        plan["resource_changes"][2]["change"]["after"]["name"] = "_Required"
        mutations.append(plan)
        plan = valid_plan()
        plan["resource_changes"][2]["change"]["after"]["disabled"] = False
        mutations.append(plan)
        for mutation in mutations:
            with self.subTest(mutation=mutation):
                self.assertTrue(MODULE.verify_plan(copy.deepcopy(mutation)))


if __name__ == "__main__":
    unittest.main()
