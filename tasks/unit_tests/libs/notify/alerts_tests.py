from __future__ import annotations

import os
import pathlib
import unittest
from collections import Counter
from contextlib import contextmanager
from unittest.mock import MagicMock, patch

from gitlab.v4.objects import ProjectJob
from invoke import MockContext, Result
from invoke.exceptions import UnexpectedExit

from tasks import notify
from tasks.libs.notify import alerts
from tasks.unit_tests.notify_tests import get_fake_jobs, get_github_slack_map


@contextmanager
def test_job_executions(path="tasks/unit_tests/testdata/job_executions.json"):
    """
    Make a job_executions.json file for testing purposes and clean it up after the test
    """
    alerts.create_initial_job_executions(path)

    yield path

    # Cancel changes
    alerts.create_initial_job_executions(path)


class TestCheckConsistentFailures(unittest.TestCase):
    @patch.dict(
        'os.environ',
        {
            'CI_PIPELINE_ID': '456',
            'CI_PIPELINE_SOURCE': 'push',
            'CI_COMMIT_BRANCH': 'taylor-swift',
            'CI_DEFAULT_BRANCH': 'taylor-swift',
        },
    )
    @patch('tasks.libs.ciproviders.gitlab_api.get_gitlab_api')
    def test_nominal(self, api_mock):
        repo_mock = api_mock.return_value.projects.get.return_value
        trace_mock = repo_mock.jobs.get.return_value.trace
        list_mock = repo_mock.pipelines.get.return_value.jobs.list

        trace_mock.return_value = b"net/http: TLS handshake timeout"
        list_mock.return_value = get_fake_jobs()

        with test_job_executions() as path:
            notify.check_consistent_failures(
                MockContext(run=Result("test")),
                1979,
                path,
            )

        repo_mock.jobs.get.assert_called()
        trace_mock.assert_called()
        list_mock.assert_called()

    @patch.dict(
        'os.environ',
        {
            'CI_PIPELINE_ID': '456',
            'CI_PIPELINE_SOURCE': 'push',
            'CI_COMMIT_BRANCH': 'taylor',
            'CI_DEFAULT_BRANCH': 'swift',
        },
    )
    @patch('tasks.libs.ciproviders.gitlab_api.get_gitlab_api')
    def test_dismiss(self, api_mock):
        repo_mock = api_mock.return_value.projects.get.return_value
        with test_job_executions() as path:
            notify.check_consistent_failures(
                MockContext(run=Result("test")),
                path,
            )
        repo_mock.jobs.get.assert_not_called()


class TestAlertsRetrieveJobExecutionsCreated(unittest.TestCase):
    job_executions = None
    job_file = "job_executions.json"

    def setUp(self) -> None:
        self.job_executions = alerts.create_initial_job_executions(self.job_file)

    def tearDown(self) -> None:
        pathlib.Path(self.job_file).unlink(missing_ok=True)

    def test_retrieved(self):
        ctx = MockContext(run=Result("test"))
        j = alerts.retrieve_job_executions(ctx, "job_executions.json")
        self.assertDictEqual(j.to_dict(), self.job_executions.to_dict())


class TestAlertsRetrieveJobExecutions(unittest.TestCase):
    test_json = "tasks/unit_tests/testdata/job_executions.json"

    def test_not_found(self):
        ctx = MagicMock()
        ctx.run.side_effect = UnexpectedExit(Result(stderr="This is a 404 not found"))
        j = alerts.retrieve_job_executions(ctx, self.test_json)
        self.assertEqual(j.pipeline_id, 0)
        self.assertEqual(j.jobs, {})

    def test_other_error(self):
        ctx = MagicMock()
        ctx.run.side_effect = UnexpectedExit(Result(stderr="This is another error"))
        with self.assertRaises(UnexpectedExit):
            alerts.retrieve_job_executions(ctx, self.test_json)


class TestAlertsUpdateStatistics(unittest.TestCase):
    @patch("tasks.libs.notify.alerts.get_failed_jobs")
    @patch("tasks.libs.notify.alerts.get_pipeline", new=MagicMock())
    def test_nominal(self, mock_get_failed):
        failed_jobs = mock_get_failed.return_value
        failed_jobs.all_failures.return_value = [
            ProjectJob(MagicMock(), attrs=a)
            for a in [{"name": "nifnif", "id": 504685380}, {"name": "nafnaf", "id": 504685380}]
        ]
        os.environ["CI_COMMIT_SHA"] = "abcdef42"
        ok = {"id": None, "failing": False, 'commit': 'abcdef42'}
        j = alerts.PipelineRuns.from_dict(
            {
                "jobs": {
                    "nafnaf": {
                        "consecutive_failures": 2,
                        "jobs_info": [
                            ok,
                            ok,
                            ok,
                            ok,
                            ok,
                            ok,
                            ok,
                            ok,
                            {"id": 422184420, "failing": True, 'commit': 'abcdef42'},
                            {"id": 618314618, "failing": True, 'commit': 'abcdef42'},
                        ],
                    },
                    "noufnouf": {
                        "consecutive_failures": 2,
                        "jobs_info": [
                            {"id": 422184420, "failing": True, 'commit': 'abcdef42'},
                            ok,
                            {"id": 618314618, "failing": True, 'commit': 'abcdef42'},
                            {"id": 314618314, "failing": True, 'commit': 'abcdef42'},
                        ],
                    },
                }
            }
        )
        a, j = alerts.update_statistics(j)
        self.assertEqual(j.jobs["nifnif"].consecutive_failures, 1)
        self.assertEqual(len(j.jobs["nifnif"].jobs_info), 1)
        self.assertTrue(j.jobs["nifnif"].jobs_info[0].failing)
        self.assertEqual(j.jobs["nafnaf"].consecutive_failures, 3)
        self.assertEqual(
            [job.failing for job in j.jobs["nafnaf"].jobs_info],
            [False, False, False, False, False, False, False, True, True, True],
        )
        self.assertEqual(j.jobs["noufnouf"].consecutive_failures, 0)
        self.assertEqual([job.failing for job in j.jobs["noufnouf"].jobs_info], [True, False, True, True, False])
        self.assertEqual(len(a["consecutive"]), 1)
        self.assertEqual(len(a["cumulative"]), 0)
        self.assertIn("nafnaf", a["consecutive"])
        mock_get_failed.assert_called()

    @patch("tasks.libs.notify.alerts.get_failed_jobs")
    @patch("tasks.libs.notify.alerts.get_pipeline", new=MagicMock())
    def test_multiple_failures(self, mock_get_failed):
        failed_jobs = mock_get_failed.return_value
        fail = {"id": 42, "failing": True, 'commit': 'abcdef42'}
        ok = {"id": None, "failing": False, 'commit': 'abcdef42'}
        failed_jobs.all_failures.return_value = [
            ProjectJob(MagicMock(), attrs=a | {"id": 42, 'commit': 'abcdef42'})
            for a in [{"name": "poulidor"}, {"name": "virenque"}, {"name": "bardet"}]
        ]
        j = alerts.PipelineRuns.from_dict(
            {
                "jobs": {
                    "poulidor": {
                        "consecutive_failures": 8,
                        "jobs_info": [ok, ok, fail, fail, fail, fail, fail, fail, fail, fail],
                    },
                    "virenque": {
                        "consecutive_failures": 2,
                        "jobs_info": [ok, ok, ok, ok, fail, ok, fail, ok, fail, fail],
                    },
                    "bardet": {"consecutive_failures": 2, "jobs_info": [fail, fail]},
                }
            }
        )
        a, j = alerts.update_statistics(j)
        self.assertEqual(j.jobs["poulidor"].consecutive_failures, 9)
        self.assertEqual(j.jobs["virenque"].consecutive_failures, 3)
        self.assertEqual(j.jobs["bardet"].consecutive_failures, 3)
        self.assertEqual(len(a["consecutive"]), 2)
        self.assertEqual(len(a["cumulative"]), 1)
        self.assertIn("virenque", a["consecutive"])
        self.assertIn("bardet", a["consecutive"])
        self.assertIn("virenque", a["cumulative"])
        mock_get_failed.assert_called()


class TestAlertsSendNotification(unittest.TestCase):
    def test_consecutive(self):
        consecutive = alerts.ConsecutiveJobAlert({'foo': [alerts.ExecutionsJobInfo(1)] * alerts.CONSECUTIVE_THRESHOLD})
        message = consecutive.format_message('abcdef')
        self.assertIn(f'{alerts.CONSECUTIVE_THRESHOLD} times in a row', message)

    def test_cumulative(self):
        cumulative = alerts.CumulativeJobAlert(
            {'foo': [alerts.ExecutionsJobInfo(i, failing=i % 3 != 0) for i in range(alerts.CUMULATIVE_LENGTH)]}
        )
        message = cumulative.message()
        self.assertIn(f'{alerts.CUMULATIVE_THRESHOLD} times in last {alerts.CUMULATIVE_LENGTH} executions', message)

    @patch('slack_sdk.WebClient', autospec=True)
    def test_none(self, mock_slack):
        client_mock = MagicMock()
        mock_slack.return_value = client_mock
        alert_jobs = {"consecutive": {}, "cumulative": {}}
        alerts.send_notification(MagicMock(), alert_jobs)
        client_mock.chat_postMessage.assert_not_called()

    @patch("tasks.libs.notify.alerts.send_metrics")
    @patch('slack_sdk.WebClient', autospec=True)
    @patch.dict('os.environ', {'SLACK_DATADOG_AGENT_BOT_TOKEN': 'coucou'})
    @patch.object(alerts.ConsecutiveJobAlert, 'message', lambda self, _: '\n'.join(self.failures) + '\n')
    @patch.object(alerts.CumulativeJobAlert, 'message', lambda self: '\n'.join(self.failures))
    @patch('tasks.owners.GITHUB_SLACK_MAP', get_github_slack_map())
    @patch('tasks.libs.notify.alerts.CHANNEL_BROADCAST', '#channel-broadcast')
    def test_jobowners(self, mock_slack: MagicMock, mock_metrics: MagicMock):
        client_mock = MagicMock()
        mock_slack.return_value = client_mock
        consecutive = {
            'tests_hello': [alerts.ExecutionsJobInfo(1)] * alerts.CONSECUTIVE_THRESHOLD,
            'tests_team_a_1': [alerts.ExecutionsJobInfo(1)] * alerts.CONSECUTIVE_THRESHOLD,
            'tests_letters_1': [alerts.ExecutionsJobInfo(1)] * alerts.CONSECUTIVE_THRESHOLD,
        }
        cumulative = {
            'tests_team_b_1': [
                alerts.ExecutionsJobInfo(i, failing=i % 3 != 0) for i in range(alerts.CUMULATIVE_LENGTH)
            ],
            'tests_team_a_2': [
                alerts.ExecutionsJobInfo(i, failing=i % 3 != 0) for i in range(alerts.CUMULATIVE_LENGTH)
            ],
        }

        alert_jobs = {"consecutive": consecutive, "cumulative": cumulative}
        alerts.send_notification(MagicMock(), alert_jobs, jobowners='tasks/unit_tests/testdata/jobowners.txt')
        self.assertEqual(len(client_mock.chat_postMessage.call_args_list), 4)

        # Verify that we send the right number of jobs per channel
        expected_team_njobs = {
            '#channel-a': 3,
            '#channel-b': 2,
            '#channel-everything': 1,
            '#channel-broadcast': 5,
        }

        for call_args in client_mock.chat_postMessage.call_args_list:
            channel, message = call_args.kwargs['channel'], call_args.kwargs['text']
            # The mock will separate job names with a newline
            jobs = message.strip().split("\n")
            njobs = len(jobs)

            self.assertEqual(
                expected_team_njobs.get(channel, None), njobs, f'Unexpected number of jobs for channel {channel}'
            )

        # Verify metrics
        mock_metrics.assert_called_once()
        expected_metrics = {
            '@datadog/team-a': 3,
            '@datadog/team-b': 2,
            '@datadog/team-everything': 1,
        }
        current_metrics = Counter()
        for metric in mock_metrics.call_args[0][0]:
            value = int(metric['points'][0]['value'])
            team = next(tag.removeprefix('team:') for tag in metric['tags'] if tag.startswith('team:'))

            current_metrics.update({team: value})

        current_metrics = dict(current_metrics.items())

        self.assertDictEqual(current_metrics, expected_metrics)

    @patch("tasks.libs.notify.alerts.send_metrics")
    @patch('slack_sdk.WebClient', autospec=True)
    @patch.dict('os.environ', {'SLACK_DATADOG_AGENT_BOT_TOKEN': 'coucou'})
    @patch.object(alerts.ConsecutiveJobAlert, 'message', lambda self, _: '\n'.join(self.failures) + '\n')
    @patch.object(alerts.CumulativeJobAlert, 'message', lambda self: '\n'.join(self.failures))
    @patch('tasks.owners.GITHUB_SLACK_MAP', get_github_slack_map())
    @patch('tasks.libs.notify.alerts.CHANNEL_BROADCAST', '#channel-a')
    def test_prevent_duplication(self, mock_slack: MagicMock, mock_metrics: MagicMock):
        client_mock = MagicMock()
        mock_slack.return_value = client_mock
        consecutive = {
            'tests_hello': [alerts.ExecutionsJobInfo(1)] * alerts.CONSECUTIVE_THRESHOLD,
            'tests_team_a_1': [alerts.ExecutionsJobInfo(1)] * alerts.CONSECUTIVE_THRESHOLD,
            'tests_letters_1': [alerts.ExecutionsJobInfo(1)] * alerts.CONSECUTIVE_THRESHOLD,
        }
        cumulative = {
            'tests_team_b_1': [
                alerts.ExecutionsJobInfo(i, failing=i % 3 != 0) for i in range(alerts.CUMULATIVE_LENGTH)
            ],
            'tests_team_a_2': [
                alerts.ExecutionsJobInfo(i, failing=i % 3 != 0) for i in range(alerts.CUMULATIVE_LENGTH)
            ],
        }

        alert_jobs = {"consecutive": consecutive, "cumulative": cumulative}
        alerts.send_notification(MagicMock(), alert_jobs, jobowners='tasks/unit_tests/testdata/jobowners.txt')
        self.assertEqual(len(client_mock.chat_postMessage.call_args_list), 3)
        mock_metrics.assert_called_once()
