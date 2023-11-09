# -*- coding: utf-8 -*-

#
# Copyright (c) 2019 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

"""
Functions and classes to simplify the execution of the `make` command.
"""

from .command import Command


class Make:
    """
    Simplifies the execution of the `make` command.
    """

    def __init__(self, log, env=None, cwd=None, common_variables={}):
        """
        Creates a new object that simplifies running the `make` command.
        """
        self._command = Command(log=log, command=["make"], env=env, cwd=cwd)
        self._common_variables = common_variables

    def _args(self, targets, variables):
        """
        Calculates the command line arguments.
        """
        result = []
        merged_variables = self._common_variables.copy()
        if variables is not None:
            merged_variables.update(variables)
        if merged_variables is not None:
            result.extend([f"{k}={v}" for k, v in merged_variables.items()])
        if targets is not None:
            result.extend(targets)
        return result

    def run(self, targets=None, variables=None):
        """
        Executes the `make` command with the given targets and variables and
        with the environment and working directory given in the constructor and
        returns its exit code.
        """
        return self._command.run(self._args(targets, variables))

    def check(self, targets=None, variables=None):
        """
        Executes the `make` command with the given targets and variables and
        with the environment and working directory given in the constructor and
        returns its exit code.
        """
        self._command.check(self._args(targets, variables))
