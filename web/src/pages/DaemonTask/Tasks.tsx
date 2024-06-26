/**
 * Copyright 2023 sigma
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import Toast from 'react-hot-toast';
import axios from "axios";
import dayjs from 'dayjs';
import parser from 'cron-parser';
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/20/solid';
import { Dialog, Listbox, Menu, Transition } from "@headlessui/react";
import { EllipsisVerticalIcon } from "@heroicons/react/20/solid";
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { Link, useLocation, useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { Tooltip } from 'flowbite';

import Header from "../../components/Header";
import IMenu from "../../components/Menu";
import Notification from "../../components/Notification";
import {
  IGcArtifactRule,
  IGcArtifactRunnerItem,
  IGcBlobRule,
  IGcBlobRunnerItem,
  IGcRepositoryRule,
  IGcRepositoryRunnerItem,
  IGcTagRule,
  IGcTagRunnerItem,
  IHTTPError
} from "../../interfaces";

const retentionAmountType = [
  { id: 1, name: 'Day' },
  { id: 2, name: 'Quantity' },
];

export default function ({ localServer }: { localServer: string }) {
  const location = useLocation();
  const navigate = useNavigate();

  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const namespaceId = searchParams.get('namespace_id') == null ? 0 : parseInt(searchParams.get('namespace_id') || "");

  const [gcRepositoryRuleExist, setGcRepositoryRuleExist] = useState(false);
  const [gcRepositoryRuleConfigModal, setGcRepositoryRuleConfigModal] = useState(false);
  const [gcRepositoryRuleRetentionDays, setGcRepositoryRuleRetentionDays] = useState<string | number>(0);
  const [gcRepositoryRuleRetentionDaysValid, setGcRepositoryRuleRetentionDaysValid] = useState(true);
  useEffect(() => { setGcRepositoryRuleRetentionDaysValid(Number.isInteger(gcRepositoryRuleRetentionDays) && parseInt(gcRepositoryRuleRetentionDays.toString()) >= 0 && parseInt(gcRepositoryRuleRetentionDays.toString()) <= 180) }, [gcRepositoryRuleRetentionDays]);
  const [gcRepositoryRuleCronEnabled, setGcRepositoryRuleCronEnabled] = useState(false);
  const [gcRepositoryRuleCronRule, setGcRepositoryRuleCronRule] = useState("");
  const [gcRepositoryRuleCronRuleValid, setGcRepositoryRuleCronRuleValid] = useState(true);
  const [gcRepositoryRuleCronRuleNextRunAt, setGcRepositoryRuleCronRuleNextRunAt] = useState("");
  const [gcRepositoryLatestRunner, setGcRepositoryLatestRunner] = useState<IGcRepositoryRunnerItem>({} as IGcRepositoryRunnerItem);
  const [gcRepositoryRule, setGcRepositoryRule] = useState<IGcRepositoryRule>({} as IGcRepositoryRule);

  const [gcTagRuleExist, setGcTagRuleExist] = useState(false);
  const [gcTagRuleConfigModal, setGcTagRuleConfigModal] = useState(false);
  const [gcTagRuleRetentionPattern, setGcTagRuleRetentionPattern] = useState("");
  const [gcTagRuleRetentionPatternValid, setGcTagRuleRetentionPatternValid] = useState(true);
  const [gcTagRuleRetentionType, setGcTagRuleRetentionType] = useState('Day');
  const [gcTagRuleRetentionAmount, setGcTagRuleRetentionAmount] = useState<string | number>(1);
  const [gcTagRuleRetentionAmountValid, setGcTagRuleRetentionAmountValid] = useState(true);
  useEffect(() => { setGcTagRuleRetentionAmountValid(Number.isInteger(gcTagRuleRetentionAmount) && parseInt(gcTagRuleRetentionAmount.toString()) >= 1 && parseInt(gcTagRuleRetentionAmount.toString()) <= 180) }, [gcTagRuleRetentionAmount]);
  const [gcTagRuleCronEnabled, setGcTagRuleCronEnabled] = useState(false);
  const [gcTagRuleCronRule, setGcTagRuleCronRule] = useState("");
  const [gcTagRuleCronRuleValid, setGcTagRuleCronRuleValid] = useState(true);
  const [gcTagRuleCronRuleNextRunAt, setGcTagRuleCronRuleNextRunAt] = useState("");
  const [gcTagLatestRunner, setGcTagLatestRunner] = useState<IGcTagRunnerItem>({} as IGcTagRunnerItem);
  const [gcTagRule, setGcTagRule] = useState<IGcTagRule>({} as IGcTagRule);

  const [gcArtifactRuleExist, setGcArtifactRuleExist] = useState(false);
  const [gcArtifactRuleConfigModal, setGcArtifactRuleConfigModal] = useState(false);
  const [gcArtifactRuleRetentionDays, setGcArtifactRuleRetentionDays] = useState<string | number>(0);
  const [gcArtifactRuleRetentionDaysValid, setGcArtifactRuleRetentionDaysValid] = useState(true);
  useEffect(() => { setGcArtifactRuleRetentionDaysValid(Number.isInteger(gcArtifactRuleRetentionDays) && parseInt(gcArtifactRuleRetentionDays.toString()) >= 0 && parseInt(gcArtifactRuleRetentionDays.toString()) <= 180) }, [gcArtifactRuleRetentionDays]);
  const [gcArtifactCronEnabled, setGcArtifactCronEnabled] = useState(false);
  const [gcArtifactRuleCronRule, setGcArtifactRuleCronRule] = useState("");
  const [gcArtifactRuleCronRuleValid, setGcArtifactRuleCronRuleValid] = useState(true);
  const [gcArtifactRuleCronRuleNextRunAt, setGcArtifactRuleCronRuleNextRunAt] = useState("");
  const [gcArtifactLatestRunner, setGcArtifactLatestRunner] = useState<IGcArtifactRunnerItem>({} as IGcArtifactRunnerItem);
  const [gcArtifactRule, setGcArtifactRule] = useState<IGcArtifactRule>({} as IGcArtifactRule);

  const [gcBlobRuleExist, setGcBlobRuleExist] = useState(false);
  const [gcBlobRuleConfigModal, setGcBlobRuleConfigModal] = useState(false);
  const [gcBlobRuleRetentionDays, setGcBlobRuleRetentionDays] = useState<string | number>(0);
  const [gcBlobRuleRetentionDaysValid, setGcBlobRuleRetentionDaysValid] = useState(true);
  useEffect(() => { setGcBlobRuleRetentionDaysValid(Number.isInteger(gcBlobRuleRetentionDays) && parseInt(gcBlobRuleRetentionDays.toString()) >= 0 && parseInt(gcBlobRuleRetentionDays.toString()) <= 180) }, [gcBlobRuleRetentionDays]);
  const [gcBlobRuleCronEnabled, setGcBlobRuleCronEnabled] = useState(false);
  const [gcBlobRuleCronRule, setGcBlobRuleCronRule] = useState("");
  const [gcBlobRuleCronRuleValid, setGcBlobRuleCronRuleValid] = useState(true);
  const [gcBlobRuleCronRuleNextRunAt, setGcBlobRuleCronRuleNextRunAt] = useState("");
  const [gcBlobLatestRunner, setGcBlobLatestRunner] = useState<IGcBlobRunnerItem>({} as IGcBlobRunnerItem);
  const [gcBlobRule, setGcBlobRule] = useState<IGcBlobRule>({} as IGcBlobRule);

  const [refreshState, setRefreshState] = useState({});

  useEffect(() => {
    const timer = setInterval(() => {
      setRefreshState({});
    }, 5000);
    return () => {
      clearInterval(timer);
    };
  }, []);

  useEffect(() => {
    let url = `${localServer}/api/v1/daemons/gc-repository/${namespaceId}/`;
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const gcRepositoryRule = response.data as IGcRepositoryRule;
        setGcRepositoryRuleExist(true);
        setGcRepositoryRule(gcRepositoryRule);
        setGcRepositoryRuleRetentionDays(gcRepositoryRule.retention_day);
        setGcRepositoryRuleCronEnabled(gcRepositoryRule.cron_enabled);
        if (gcRepositoryRule.cron_enabled) {
          setGcRepositoryRuleCronRule(gcRepositoryRule.cron_rule == undefined ? "" : gcRepositoryRule.cron_rule);
        }
        let url = `${localServer}/api/v1/daemons/gc-repository/${namespaceId}/runners/latest`;
        axios.get(url).then(response => {
          if (response?.status === 200) {
            const runner = response.data as IGcRepositoryRunnerItem;
            setGcRepositoryLatestRunner(runner);
          } else if (response?.status === 404) {
            // do nothing
          } else {
            const errorcode = response.data as IHTTPError;
            Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
          }
        }).catch(error => {
          const errorcode = error.response.data as IHTTPError;
          Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
        })
      } else if (response?.status === 404) {
        setGcRepositoryRuleExist(false);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    })
  }, [refreshState]);

  useEffect(() => {
    let url = `${localServer}/api/v1/daemons/gc-tag/${namespaceId}/`;
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const gcTagRule = response.data as IGcTagRule;
        setGcTagRuleExist(true);
        setGcTagRule(gcTagRule);
        setGcTagRuleCronEnabled(gcTagRule.cron_enabled);
        setGcTagRuleRetentionType(gcTagRule.retention_rule_type);
        setGcTagRuleRetentionAmount(gcTagRule.retention_rule_amount);
        if (gcTagRule.cron_enabled) {
          setGcTagRuleCronRule(gcTagRule.cron_rule == undefined ? "" : gcTagRule.cron_rule)
        }
        if (gcTagRule.retention_pattern != undefined) {
          setGcTagRuleRetentionPattern(gcTagRule.retention_pattern == undefined ? "" : gcTagRule.retention_pattern);
        }
        let url = `${localServer}/api/v1/daemons/gc-tag/${namespaceId}/runners/latest`;
        axios.get(url).then(response => {
          if (response?.status === 200) {
            const runner = response.data as IGcTagRunnerItem;
            setGcTagLatestRunner(runner);
          } else if (response?.status === 404) {
            // do nothing
          } else {
            const errorcode = response.data as IHTTPError;
            Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
          }
        }).catch(error => {
          const errorcode = error.response.data as IHTTPError;
          Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
        })
      } else if (response?.status === 404) {
        setGcTagRuleExist(false);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    })
  }, [refreshState]);

  useEffect(() => {
    let url = `${localServer}/api/v1/daemons/gc-artifact/${namespaceId}/`;
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const gcArtifactRule = response.data as IGcArtifactRule;
        setGcArtifactRuleExist(true);
        setGcArtifactRule(gcArtifactRule);
        setGcArtifactCronEnabled(gcArtifactRule.cron_enabled);
        if (gcArtifactRule.cron_enabled) {
          setGcArtifactRuleCronRule(gcArtifactRule.cron_rule == undefined ? "" : gcArtifactRule.cron_rule);
        }
        setGcArtifactRuleRetentionDays(gcArtifactRule.retention_day);
        let url = `${localServer}/api/v1/daemons/gc-artifact/${namespaceId}/runners/latest`;
        axios.get(url).then(response => {
          if (response?.status === 200) {
            const runner = response.data as IGcArtifactRunnerItem;
            setGcArtifactLatestRunner(runner);
          } else if (response?.status === 404) {
            // do nothing
          } else {
            const errorcode = response.data as IHTTPError;
            Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
          }
        }).catch(error => {
          const errorcode = error.response.data as IHTTPError;
          Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
        })
      } else if (response?.status === 404) {
        setGcArtifactRuleExist(false);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    })
  }, [refreshState]);

  useEffect(() => {
    if (!location.pathname.startsWith("/settings")) {
      return;
    }
    let url = `${localServer}/api/v1/daemons/gc-blob/${namespaceId}/`;
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const gcBlobRule = response.data as IGcBlobRule;
        setGcBlobRuleExist(true);
        setGcBlobRule(gcBlobRule);
        setGcBlobRuleCronEnabled(gcBlobRule.cron_enabled);
        setGcBlobRuleRetentionDays(gcBlobRule.retention_day);
        if (gcBlobRule.cron_enabled) {
          setGcBlobRuleCronRule(gcBlobRule.cron_rule == undefined ? "" : gcBlobRule.cron_rule)
        }
        let url = `${localServer}/api/v1/daemons/gc-blob/${namespaceId}/runners/latest`;
        axios.get(url).then(response => {
          if (response?.status === 200) {
            const runner = response.data as IGcBlobRunnerItem;
            setGcBlobLatestRunner(runner);
          } else if (response?.status === 404) {
            // do nothing
          } else {
            const errorcode = response.data as IHTTPError;
            Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
          }
        }).catch(error => {
          const errorcode = error.response.data as IHTTPError;
          Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
        })
      } else if (response?.status === 404) {
        setGcBlobRuleExist(false);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    })
  }, [refreshState]);

  useEffect(() => {
    if (gcArtifactCronEnabled && gcArtifactRuleCronRule.length > 0) {
      axios.post(localServer + `/api/v1/validators/cron`, {
        cron: gcArtifactRuleCronRule,
      }).then(response => {
        if (response?.status === 204) {
          setGcArtifactRuleCronRuleValid(true);
          let next = parser.parseExpression(gcArtifactRuleCronRule).next()
          setGcArtifactRuleCronRuleNextRunAt(`${dayjs(next.toDate()).format('YYYY-MM-DD HH:mm')}`);
        } else {
          setGcArtifactRuleCronRuleValid(false);
        }
      }).catch(error => {
        console.log(error);
        setGcArtifactRuleCronRuleValid(false);
      });
    }
  }, [gcArtifactRuleCronRule, gcArtifactCronEnabled]);

  useEffect(() => {
    if (gcTagRuleRetentionPattern.length > 0) {
      axios.post(localServer + `/api/v1/validators/regexp`, {
        regexp: gcTagRuleRetentionPattern,
      }).then(response => {
        if (response?.status === 204) {
          setGcTagRuleRetentionPatternValid(true);
        } else {
          setGcTagRuleRetentionPatternValid(false);
        }
      }).catch(error => {
        console.log(error);
        setGcTagRuleRetentionPatternValid(false);
      });
    }
  }, [gcTagRuleRetentionPattern]);

  useEffect(() => {
    if (gcRepositoryRuleCronEnabled && gcRepositoryRuleCronRule.length > 0) {
      axios.post(localServer + `/api/v1/validators/cron`, {
        cron: gcRepositoryRuleCronRule,
      }).then(response => {
        if (response?.status === 204) {
          setGcRepositoryRuleCronRuleValid(true);
          let next = parser.parseExpression(gcRepositoryRuleCronRule).next()
          setGcRepositoryRuleCronRuleNextRunAt(`${dayjs(next.toDate()).format('YYYY-MM-DD HH:mm')}`);
        } else {
          setGcRepositoryRuleCronRuleValid(false);
        }
      }).catch(error => {
        console.log(error);
        setGcRepositoryRuleCronRuleValid(false);
      });
    }
  }, [gcRepositoryRuleCronRule, gcRepositoryRuleCronEnabled]);

  useEffect(() => {
    if (gcBlobRuleCronEnabled && gcBlobRuleCronRule.length > 0) {
      axios.post(localServer + `/api/v1/validators/cron`, {
        cron: gcBlobRuleCronRule,
      }).then(response => {
        if (response?.status === 204) {
          setGcBlobRuleCronRuleValid(true);
          let next = parser.parseExpression(gcBlobRuleCronRule).next()
          setGcBlobRuleCronRuleNextRunAt(`${dayjs(next.toDate()).format('YYYY-MM-DD HH:mm')}`);
        } else {
          setGcBlobRuleCronRuleValid(false);
        }
      }).catch(error => {
        console.log(error);
        setGcBlobRuleCronRuleValid(false);
      });
    }
  }, [gcBlobRuleCronRule, gcBlobRuleCronEnabled]);

  useEffect(() => {
    if (gcTagRuleCronEnabled && gcTagRuleCronRule.length > 0) {
      axios.post(localServer + `/api/v1/validators/cron`, {
        cron: gcTagRuleCronRule,
      }).then(response => {
        if (response?.status === 204) {
          setGcTagRuleCronRuleValid(true);
          let next = parser.parseExpression(gcTagRuleCronRule).next()
          setGcTagRuleCronRuleNextRunAt(`${dayjs(next.toDate()).format('YYYY-MM-DD HH:mm')}`);
        } else {
          setGcTagRuleCronRuleValid(false);
        }
      }).catch(error => {
        console.log(error);
        setGcTagRuleCronRuleValid(false);
      });
    }
  }, [gcTagRuleCronRule, gcTagRuleCronEnabled]);

  const createOrUpdateGcRepository = () => {
    if (!(gcRepositoryRuleRetentionDaysValid && ((gcRepositoryRuleCronEnabled && gcRepositoryRuleCronRuleValid) || !gcRepositoryRuleCronEnabled))) {
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    if (gcRepositoryRuleCronEnabled && gcRepositoryRuleCronRule == "") {
      setGcRepositoryRuleCronRuleValid(false);
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    const data: { [key: string]: any } = {
      retention_day: gcRepositoryRuleRetentionDays,
      cron_enabled: gcRepositoryRuleCronEnabled,
    };
    if (gcRepositoryRuleCronEnabled) {
      data["cron_rule"] = gcRepositoryRuleCronRule;
    }
    axios.put(localServer + `/api/v1/daemons/gc-repository/${namespaceId}/`, data).then(response => {
      if (response?.status === 204) {
        let message = "Create garbage collect empty repository config success"
        if (gcRepositoryRuleExist) {
          message = "Update garbage collect empty repository config success"
        }
        Toast.success(message);
        setGcRepositoryRuleConfigModal(false);
        setRefreshState({});
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  const createOrUpdateGcArtifact = () => {
    if (!(gcArtifactRuleRetentionDaysValid && ((gcArtifactCronEnabled && gcArtifactRuleCronRuleValid) || !gcArtifactCronEnabled))) {
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    if (gcArtifactCronEnabled && gcArtifactRuleCronRule == "") {
      setGcArtifactRuleCronRuleValid(false);
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    const data: { [key: string]: any } = {
      retention_day: gcArtifactRuleRetentionDays,
      cron_enabled: gcArtifactCronEnabled,
    };
    if (gcArtifactCronEnabled) {
      data["cron_rule"] = gcArtifactRuleCronRule;
    }
    axios.put(localServer + `/api/v1/daemons/gc-artifact/${namespaceId}/`, data).then(response => {
      if (response?.status === 204) {
        let message = "Create garbage collect artifact config success"
        if (gcArtifactRuleExist) {
          message = "Update garbage collect artifact config success"
        }
        Toast.success(message);
        setGcArtifactRuleConfigModal(false);
        setRefreshState({});
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  const createOrUpdateGcBlob = () => {
    if (!(gcBlobRuleRetentionDaysValid && ((gcBlobRuleCronEnabled && gcBlobRuleCronRuleValid) || !gcBlobRuleCronEnabled))) {
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    if (gcBlobRuleCronEnabled && gcBlobRuleCronRule == "") {
      setGcBlobRuleCronRuleValid(false);
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    const data: { [key: string]: any } = {
      retention_day: gcBlobRuleRetentionDays,
      cron_enabled: gcBlobRuleCronEnabled,
    };
    if (gcBlobRuleCronEnabled) {
      data["cron_rule"] = gcBlobRuleCronRule;
    }
    axios.put(localServer + `/api/v1/daemons/gc-blob/${namespaceId}/`, data).then(response => {
      if (response?.status === 204) {
        let message = "Create garbage collect blob config success";
        if (gcBlobRuleExist) {
          message = "Update garbage collect blob config success";
        }
        Toast.success(message);
        setGcBlobRuleConfigModal(false);
        setRefreshState({});
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  const createOrUpdateGcTag = () => {
    if (!(gcTagRuleRetentionAmountValid && ((gcTagRuleCronEnabled && gcTagRuleCronRuleValid) || !gcTagRuleCronEnabled))) {
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    if (gcTagRuleCronEnabled && gcTagRuleCronRule == "") {
      setGcTagRuleCronRuleValid(false);
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    if (gcTagRuleRetentionPattern != "" && !gcTagRuleRetentionPatternValid) {
      setGcTagRuleRetentionPatternValid(false);
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    const data: { [key: string]: any } = {
      retention_rule_type: gcTagRuleRetentionType,
      retention_rule_amount: gcTagRuleRetentionAmount,
      cron_enabled: gcTagRuleCronEnabled,
    };
    if (gcTagRuleCronEnabled) {
      data["cron_rule"] = gcTagRuleCronRule;
    }
    if (gcTagRuleRetentionPattern != "") {
      data["retention_pattern"] = gcTagRuleRetentionPattern;
    }
    axios.put(localServer + `/api/v1/daemons/gc-tag/${namespaceId}/`, data).then(response => {
      if (response?.status === 204) {
        let message = "Create garbage collect tag config success"
        if (gcTagRuleExist) {
          message = "Update garbage collect tag config success"
        }
        Toast.success(message);
        setGcTagRuleConfigModal(false);
        setRefreshState({});
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  const createGcRepositoryRunner = () => {
    axios.post(localServer + `/api/v1/daemons/gc-repository/${namespaceId}/runners/`, {}).then(response => {
      if (response?.status === 201) {
        Toast.success("Garbage collect repository task will run in seconds");
        setRefreshState({});
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  const createGcTagRunner = () => {
    axios.post(localServer + `/api/v1/daemons/gc-tag/${namespaceId}/runners/`, {}).then(response => {
      if (response?.status === 201) {
        Toast.success("Garbage collect tag task will run in seconds");
        setRefreshState({});
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  const createGcArtifactRunner = () => {
    axios.post(localServer + `/api/v1/daemons/gc-artifact/${namespaceId}/runners/`, {}).then(response => {
      if (response?.status === 201) {
        Toast.success("Garbage collect artifact task will run in seconds");
        setRefreshState({});
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  const createGcBlobRunner = () => {
    axios.post(localServer + `/api/v1/daemons/gc-blob/${namespaceId}/runners/`, {}).then(response => {
      if (response?.status === 201) {
        Toast.success("Garbage collect blob task will run in seconds");
        setRefreshState({});
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Namespace Daemon Task</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-white">
        <IMenu localServer={localServer} item={location.pathname.startsWith("/settings") ? "daemon-tasks" : "repositories"} namespace={namespace} />
        <div className="flex flex-col w-0 flex-1 overflow-visible">
          <main className="relative z-0 focus:outline-none" tabIndex={0}>
            <Header title={location.pathname.startsWith("/settings") ? "Setting - Daemon Task" : "Namespace - Daemon Task"}
              props={
                location.pathname.startsWith("/settings") ? null : (
                  <div className="flex space-x-8">
                    <Link
                      to={`/namespaces/${namespace}/namespace-summary?namespace_id=${namespaceId}`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Summary
                    </Link>
                    <Link
                      to={`/namespaces/${namespace}/repositories?namespace_id=${namespaceId}`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Repository list
                    </Link>
                    <Link
                      to={`/namespaces/${namespace}/members?namespace_id=${namespaceId}`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Members
                    </Link>
                    <Link
                      to={`/namespaces/${namespace}/daemon-tasks?namespace_id=${namespaceId}`}
                      className="inline-flex items-center border-b border-indigo-500 px-1 pt-1 text-sm font-medium text-gray-900 capitalize"
                      onClick={e => {
                        e.preventDefault();
                      }}
                    >
                      Daemon task
                    </Link>
                    <Link
                      to={`/namespaces/${namespace}/webhooks?namespace_id=${namespaceId}`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Webhook
                    </Link>
                  </div>
                )
              } />
            <div className="flex flex-1 overflow-visible">
              <div className="align-middle inline-block min-w-full border-gray-200">
                <table className="min-w-full flex-1 overflow-visible">
                  <thead>
                    <tr className="border-gray-200">
                      <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                        <span className="lg:pl-2">Task</span>
                      </th>
                      <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                        <span className="lg:pl-2">Running</span>
                      </th>
                      <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                        <span className="lg:pl-2">Last trigger</span>
                      </th>
                      <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                        <span className="lg:pl-2">Next trigger</span>
                      </th>
                      <th className="sticky top-0 z-10 pr-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                        <span className="lg:pl-2">Action</span>
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr className="border-b">
                      <td className="px-6 py-4 max-w-0 w-full whitespace-nowrap text-sm font-normal text-gray-900 cursor-pointer"
                        onClick={() => {
                          if (!gcRepositoryRuleExist) {
                            return;
                          }
                          if (location.pathname.startsWith("/settings")) {
                            navigate(`/settings/daemon-tasks/gc-repository?namespace_id=${namespaceId}`);
                          } else {
                            navigate(`/namespaces/${namespace}/daemon-tasks/gc-repository?namespace_id=${namespaceId}`);
                          }
                        }}
                      >
                        <div className="flex items-center space-x-3 lg:pl-2">
                          Garbage collect the empty repositories
                        </div>
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        {gcRepositoryRuleExist && gcRepositoryRule.is_running ? "true" : "false"}
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        {gcRepositoryRuleExist && gcRepositoryLatestRunner.started_at != undefined ? gcRepositoryLatestRunner.started_at : "-"}
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        {gcRepositoryRuleExist && gcRepositoryRule.cron_enabled && gcRepositoryRule.cron_next_trigger != undefined ? gcRepositoryRule.cron_next_trigger : "-"}
                      </td>
                      <td className="pr-3 whitespace-nowrap">
                        <Menu as="div" className="relative flex-none" onClick={e => {
                          e.stopPropagation();
                        }}>
                          <Menu.Button className="mx-auto -m-2.5 block p-2.5 text-gray-500 hover:text-gray-900 margin">
                            <span className="sr-only">Open options</span>
                            <EllipsisVerticalIcon className="h-5 w-5" aria-hidden="true" />
                          </Menu.Button>
                          <Transition
                            as={Fragment}
                            enter="transition ease-out duration-100"
                            enterFrom="transform opacity-0 scale-95"
                            enterTo="transform opacity-100 scale-100"
                            leave="transition ease-in duration-75"
                            leaveFrom="transform opacity-100 scale-100"
                            leaveTo="transform opacity-0 scale-95"
                          >
                            <Menu.Items className={(1 > 10 ? "menu-action-top" : "mt-2") + " text-left absolute right-0 z-10 w-30 origin-top-right rounded-md bg-white py-2 shadow-lg ring-1 ring-gray-900/5 focus:outline-none"} >
                              <Menu.Item>
                                {({ active }) => (
                                  <div
                                    className={
                                      (active ? 'bg-gray-100' : '') +
                                      ' block px-3 py-1 text-sm leading-6 text-gray-900 cursor-pointer'
                                    }
                                    onClick={e => { setGcRepositoryRuleConfigModal(true); }}
                                  >
                                    {
                                      gcRepositoryRuleExist ? "Update" : "Configuration"
                                    }
                                  </div>
                                )}
                              </Menu.Item>
                              <Menu.Item
                                disabled={!gcRepositoryRuleExist}
                              >
                                {({ active }) => (
                                  <div
                                    className={
                                      (active ? 'bg-gray-100' : '') +
                                      (gcRepositoryRuleExist ? ' cursor-pointer ' : ' cursor-not-allowed ') +
                                      ' block px-3 py-1 text-sm leading-6 text-gray-900'
                                    }
                                    onClick={e => {
                                      createGcRepositoryRunner();
                                    }}
                                  >
                                    Run
                                  </div>
                                )}
                              </Menu.Item>
                            </Menu.Items>
                          </Transition>
                        </Menu>
                      </td>
                    </tr>
                    <tr className="border-b">
                      <td className="px-6 py-4 max-w-0 w-full whitespace-nowrap text-sm font-normal text-gray-900 cursor-pointer"
                        onClick={() => {
                          if (!gcTagRuleExist) {
                            return;
                          }
                          if (location.pathname.startsWith("/settings")) {
                            navigate(`/settings/daemon-tasks/gc-tag?namespace_id=${namespaceId}`);
                          } else {
                            navigate(`/namespaces/${namespace}/daemon-tasks/gc-tag?namespace_id=${namespaceId}`);
                          }
                        }}
                      >
                        <div className="flex items-center space-x-3 lg:pl-2">
                          Garbage collect the tags
                        </div>
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        {gcTagRuleExist && gcTagRule.is_running ? "true" : "false"}
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        {gcTagRuleExist && gcTagLatestRunner.started_at != undefined ? gcTagLatestRunner.started_at : "-"}
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        {gcTagRuleExist && gcTagRule.cron_enabled && gcTagRule.cron_next_trigger != undefined ? gcTagRule.cron_next_trigger : "-"}
                      </td>
                      <td className="pr-3 whitespace-nowrap">
                        <Menu as="div" className="relative flex-none" onClick={e => {
                          e.stopPropagation();
                        }}>
                          <Menu.Button className="mx-auto -m-2.5 block p-2.5 text-gray-500 hover:text-gray-900 margin">
                            <span className="sr-only">Open options</span>
                            <EllipsisVerticalIcon className="h-5 w-5" aria-hidden="true" />
                          </Menu.Button>
                          <Transition
                            as={Fragment}
                            enter="transition ease-out duration-100"
                            enterFrom="transform opacity-0 scale-95"
                            enterTo="transform opacity-100 scale-100"
                            leave="transition ease-in duration-75"
                            leaveFrom="transform opacity-100 scale-100"
                            leaveTo="transform opacity-0 scale-95"
                          >
                            <Menu.Items className={(1 > 10 ? "menu-action-top" : "mt-2") + " text-left absolute right-0 z-10 w-30 origin-top-right rounded-md bg-white py-2 shadow-lg ring-1 ring-gray-900/5 focus:outline-none"} >
                              <Menu.Item>
                                {({ active }) => (
                                  <div
                                    className={
                                      (active ? 'bg-gray-100' : '') +
                                      ' block px-3 py-1 text-sm leading-6 text-gray-900 cursor-pointer'
                                    }
                                    onClick={e => { setGcTagRuleConfigModal(true) }}
                                  >
                                    {
                                      gcTagRuleExist ? "Update" : "Configuration"
                                    }
                                  </div>
                                )}
                              </Menu.Item>
                              <Menu.Item
                                disabled={!gcTagRuleExist}
                              >
                                {({ active }) => (
                                  <div
                                    className={
                                      (active ? 'bg-gray-100' : '') +
                                      (gcTagRuleExist ? ' cursor-pointer ' : ' cursor-not-allowed ') +
                                      ' block px-3 py-1 text-sm leading-6 text-gray-900'
                                    }
                                    onClick={e => createGcTagRunner()}
                                  >
                                    Run
                                  </div>
                                )}
                              </Menu.Item>
                            </Menu.Items>
                          </Transition>
                        </Menu>
                      </td>
                    </tr>
                    <tr className="border-b">
                      <td className="px-6 py-4 max-w-0 w-full whitespace-nowrap text-sm font-normal text-gray-900 cursor-pointer"
                        onClick={() => {
                          if (!gcArtifactRuleExist) {
                            return;
                          }
                          if (location.pathname.startsWith("/settings")) {
                            navigate(`/settings/daemon-tasks/gc-artifact?namespace_id=${namespaceId}`);
                          } else {
                            navigate(`/namespaces/${namespace}/daemon-tasks/gc-artifact?namespace_id=${namespaceId}`);
                          }
                        }}
                      >
                        <div className="flex items-center space-x-3 lg:pl-2">
                          Garbage collect the artifacts
                        </div>
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        {gcArtifactRuleExist && gcArtifactRule.is_running ? "true" : "false"}
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        {gcArtifactRuleExist && gcArtifactLatestRunner.started_at != undefined ? gcArtifactLatestRunner.started_at : "-"}
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        {gcArtifactRuleExist && gcArtifactRule.cron_enabled && gcArtifactRule.cron_next_trigger != undefined ? gcArtifactRule.cron_next_trigger : "-"}
                      </td>
                      <td className="pr-3 whitespace-nowrap">
                        <Menu as="div" className="relative flex-none" onClick={e => {
                          e.stopPropagation();
                        }}>
                          <Menu.Button className="mx-auto -m-2.5 block p-2.5 text-gray-500 hover:text-gray-900 margin">
                            <span className="sr-only">Open options</span>
                            <EllipsisVerticalIcon className="h-5 w-5" aria-hidden="true" />
                          </Menu.Button>
                          <Transition
                            as={Fragment}
                            enter="transition ease-out duration-100"
                            enterFrom="transform opacity-0 scale-95"
                            enterTo="transform opacity-100 scale-100"
                            leave="transition ease-in duration-75"
                            leaveFrom="transform opacity-100 scale-100"
                            leaveTo="transform opacity-0 scale-95"
                          >
                            <Menu.Items className={(1 > 10 ? "menu-action-top" : "mt-2") + " text-left absolute right-0 z-10 w-30 origin-top-right rounded-md bg-white py-2 shadow-lg ring-1 ring-gray-900/5 focus:outline-none"} >
                              <Menu.Item>
                                {({ active }) => (
                                  <div
                                    className={
                                      (active ? 'bg-gray-100' : '') +
                                      ' block px-3 py-1 text-sm leading-6 text-gray-900 cursor-pointer'
                                    }
                                    onClick={e => { setGcArtifactRuleConfigModal(true) }}
                                  >
                                    {
                                      gcArtifactRuleExist ? "Update" : "Configuration"
                                    }
                                  </div>
                                )}
                              </Menu.Item>
                              <Menu.Item
                                disabled={!gcArtifactRuleExist}
                              >
                                {({ active }) => (
                                  <div
                                    className={
                                      (active ? 'bg-gray-100' : '') +
                                      (gcArtifactRuleExist ? ' cursor-pointer ' : ' cursor-not-allowed ') +
                                      ' block px-3 py-1 text-sm leading-6 text-gray-900'
                                    }
                                    onClick={e => createGcArtifactRunner()}
                                  >
                                    Run
                                  </div>
                                )}
                              </Menu.Item>
                            </Menu.Items>
                          </Transition>
                        </Menu>
                      </td>
                    </tr>
                    {
                      namespaceId !== 0 ? null : (
                        <tr className="border-b">
                          <td className="px-6 py-4 max-w-0 w-full whitespace-nowrap text-sm font-normal text-gray-900 cursor-pointer"
                            onClick={() => {
                              if (!gcBlobRuleExist) {
                                return;
                              }
                              if (location.pathname.startsWith("/settings")) {
                                navigate(`/settings/daemon-tasks/gc-blob?namespace_id=${namespaceId}`);
                              } else {
                                navigate(`/namespaces/${namespace}/daemon-tasks/gc-blob?namespace_id=${namespaceId}`);
                              }
                            }}
                          >
                            <div className="flex items-center space-x-3 lg:pl-2">
                              Garbage collect the blobs
                            </div>
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {gcBlobRuleExist && gcBlobRule.is_running ? "true" : "false"}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {gcBlobRuleExist && gcBlobLatestRunner.started_at != undefined ? gcBlobLatestRunner.started_at : "-"}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {gcBlobRuleExist && gcBlobRule.cron_enabled && gcBlobRule.cron_next_trigger != undefined ? gcBlobRule.cron_next_trigger : "-"}
                          </td>
                          <td className="pr-3 whitespace-nowrap">
                            <Menu as="div" className="relative flex-none" onClick={e => {
                              e.stopPropagation();
                            }}>
                              <Menu.Button className="mx-auto -m-2.5 block p-2.5 text-gray-500 hover:text-gray-900 margin">
                                <span className="sr-only">Open options</span>
                                <EllipsisVerticalIcon className="h-5 w-5" aria-hidden="true" />
                              </Menu.Button>
                              <Transition
                                as={Fragment}
                                enter="transition ease-out duration-100"
                                enterFrom="transform opacity-0 scale-95"
                                enterTo="transform opacity-100 scale-100"
                                leave="transition ease-in duration-75"
                                leaveFrom="transform opacity-100 scale-100"
                                leaveTo="transform opacity-0 scale-95"
                              >
                                <Menu.Items className={(1 > 10 ? "menu-action-top" : "mt-2") + " text-left absolute right-0 z-10 w-30 origin-top-right rounded-md bg-white py-2 shadow-lg ring-1 ring-gray-900/5 focus:outline-none"} >
                                  <Menu.Item>
                                    {({ active }) => (
                                      <div
                                        className={
                                          (active ? 'bg-gray-100' : '') +
                                          ' block px-3 py-1 text-sm leading-6 text-gray-900 cursor-pointer'
                                        }
                                        onClick={e => { setGcBlobRuleConfigModal(true) }}
                                      >
                                        {
                                          gcBlobRuleExist ? "Update" : "Configuration"
                                        }
                                      </div>
                                    )}
                                  </Menu.Item>
                                  <Menu.Item
                                    disabled={!gcBlobRuleExist}
                                  >
                                    {({ active }) => (
                                      <div
                                        className={
                                          (active ? 'bg-gray-100' : '') +
                                          (gcBlobRuleExist ? ' cursor-pointer ' : ' cursor-not-allowed ') +
                                          ' block px-3 py-1 text-sm leading-6 text-gray-900'
                                        }
                                        onClick={e => createGcBlobRunner()}
                                      >
                                        Run
                                      </div>
                                    )}
                                  </Menu.Item>
                                </Menu.Items>
                              </Transition>
                            </Menu>
                          </td>
                        </tr>
                      )
                    }
                  </tbody>
                </table>
              </div>
            </div>
          </main>
        </div>
      </div>
      <div
        id="tooltip-gc-repository-retention-days"
        role="tooltip"
        className="absolute z-50 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700 w-[350px]">
        Retention the empty repository for specific days,
        0 means delete immediately, available 0-180
        <div className="tooltip-arrow" data-popper-arrow></div>
      </div>
      <div
        id="tooltip-gc-repository-cron-rule"
        role="tooltip"
        className="absolute z-50 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700">
        '0 0 * * 6' means run at 00:00 every Saturday
        <div className="tooltip-arrow" data-popper-arrow></div>
      </div>
      <div
        id="tooltip-gc-blob-retention-days"
        role="tooltip"
        className="absolute z-50 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700 w-[350px]">
        Retention the blob for specific days,
        0 means delete immediately, available 0-180
        <div className="tooltip-arrow" data-popper-arrow></div>
      </div>
      <div
        id="tooltip-gc-blob-cron-rule"
        role="tooltip"
        className="absolute z-50 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700">
        '0 0 * * 6' means run at 00:00 every Saturday
        <div className="tooltip-arrow" data-popper-arrow></div>
      </div>
      <div
        id="tooltip-gc-artifact-retention-days"
        role="tooltip"
        className="absolute z-50 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700 w-[350px]">
        Retention the artifact for specific days,
        0 means delete immediately, available 0-180
        <div className="tooltip-arrow" data-popper-arrow></div>
      </div>
      <div
        id="tooltip-gc-artifact-cron-rule"
        role="tooltip"
        className="absolute z-50 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700">
        '0 0 * * 6' means run at 00:00 every Saturday
        <div className="tooltip-arrow" data-popper-arrow></div>
      </div>
      <div
        id="tooltip-gc-tag-retention-amount"
        role="tooltip"
        className="absolute z-50 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700 w-[350px]">
        Retention the tag for specific days or quantity,
        available 1-180
        <div className="tooltip-arrow" data-popper-arrow></div>
      </div>
      <div
        id="tooltip-gc-tag-retention-pattern"
        role="tooltip"
        className="absolute z-50 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700 w-[350px]">
        Retention the tag for specific regexp,
        please try 'v.*'
        <div className="tooltip-arrow" data-popper-arrow></div>
      </div>
      <div
        id="tooltip-gc-tag-cron-rule"
        role="tooltip"
        className="absolute z-50 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700">
        '0 0 * * 6' means run at 00:00 every Saturday
        <div className="tooltip-arrow" data-popper-arrow></div>
      </div>
      <Transition.Root show={gcBlobRuleConfigModal} as={Fragment}>
        <Dialog as="div" className="relative z-10" onClose={setGcBlobRuleConfigModal}>
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
          </Transition.Child>
          <div className="fixed inset-0 z-10 overflow-y-auto">
            <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
              <Transition.Child
                as={Fragment}
                enter="ease-out duration-300"
                enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                enterTo="opacity-100 translate-y-0 sm:scale-100"
                leave="ease-in duration-200"
                leaveFrom="opacity-100 translate-y-0 sm:scale-100"
                leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              >
                <Dialog.Panel className="relative transform overflow-hidden rounded-lg bg-white px-6 pb-4 text-left shadow-xl transition-all">
                  <Dialog.Title
                    as="h3"
                    className="text-lg font-medium leading-6 text-gray-900 border-b pt-4 pb-4"
                  >
                    Garbage collect blob config
                  </Dialog.Title>
                  <div className="flex flex-col gap-0 mt-4">
                    <div className="grid grid-cols-6 gap-4">
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            <span className="text-red-600">*</span>
                            <span className="leading-6 ">Retention Days</span>
                            <div className="flex flex-row cursor-pointer"
                              id="gcBlobRetentionDaysHelp"
                              onClick={e => {
                                let tooltip = new Tooltip(document.getElementById("tooltip-gc-blob-retention-days"),
                                  document.getElementById("gcBlobRetentionDaysHelp"), { triggerType: "click" });
                                tooltip.show();
                              }}
                            >
                              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 block my-auto ml-0.5">
                                <path strokeLinecap="round" strokeLinejoin="round" d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z" />
                              </svg>
                            </div>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-4">
                        <div className="relative rounded-md shadow-sm">
                          <input
                            type="text"
                            id="namespace_count_limit"
                            name="namespace_count_limit"
                            placeholder="0 means no limit"
                            className={(gcBlobRuleRetentionDaysValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                            value={gcBlobRuleRetentionDays}
                            onChange={e => setGcBlobRuleRetentionDays(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))}
                          />
                          {
                            gcBlobRuleRetentionDaysValid ? null : (
                              <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-red-500">
                                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                                </svg>
                              </div>
                            )
                          }
                        </div>
                      </div>
                    </div>
                    <div className="grid grid-cols-6 gap-4">
                      <div className="col-span-2"></div>
                      <div className="col-span-4">
                        {
                          gcBlobRuleRetentionDaysValid ? null : (
                            <p className="mt-1 text-xs text-red-600">
                              <span>
                                Not a valid retention days limit, available 0-180.
                              </span>
                            </p>
                          )
                        }
                      </div>
                    </div>
                    <div className="grid grid-cols-6 gap-4 mt-4">
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            <span className="text-red-600">*</span>
                            <span className="leading-6 ">Cron Enabled</span>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-4">
                        <div className="mt-0.5 flex flex-row items-center h-[36px]">
                          <label className="relative inline-flex items-center cursor-pointer">
                            <input type="checkbox" checked={gcBlobRuleCronEnabled} className="sr-only peer" onChange={e => {
                              setGcBlobRuleCronEnabled(e.target.checked);
                            }} />
                            <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                          </label>
                        </div>
                      </div>
                    </div>
                    {
                      !gcBlobRuleCronEnabled ? null : (
                        <>
                          <div className="grid grid-cols-6 gap-4 mt-4">
                            <div className="col-span-2 flex flex-row">
                              <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                                <div className="flex">
                                  <span className="text-red-600">*</span>
                                  <span className="leading-6 ">Cron Rule</span>
                                  <div className="flex flex-row cursor-pointer"
                                    id="gcBlobRuleHelp"
                                    onClick={e => {
                                      let tooltip = new Tooltip(document.getElementById("tooltip-gc-blob-cron-rule"),
                                        document.getElementById("gcBlobRuleHelp"), { triggerType: "click" });
                                      tooltip.show();
                                    }}
                                  >
                                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 block my-auto ml-0.5">
                                      <path strokeLinecap="round" strokeLinejoin="round" d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z" />
                                    </svg>
                                  </div>
                                  <span>:</span>
                                </div>
                              </label>
                            </div>

                            <div className="col-span-4">
                              <div className="relative rounded-md shadow-sm">
                                <input
                                  type="text"
                                  id="gc_repository_cron_rule"
                                  name="gc_repository_cron_rule"
                                  placeholder="cron rule"
                                  className={(gcBlobRuleCronRuleValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                                  value={gcBlobRuleCronRule}
                                  onChange={e => setGcBlobRuleCronRule(e.target.value)}
                                />
                                {
                                  gcBlobRuleCronRuleValid ? null : (
                                    <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-red-500">
                                        <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                                      </svg>
                                    </div>
                                  )
                                }
                              </div>
                            </div>
                          </div>
                          <div className="grid grid-cols-6 gap-4">
                            <div className="col-span-2">
                            </div>
                            <div className="col-span-4">
                              {
                                !gcBlobRuleCronRuleValid ? (
                                  <p className="mt-1 text-xs text-red-600">
                                    <span>
                                      Not a valid cron rule, you can try '0 0 * * 6'.
                                    </span>
                                  </p>
                                ) : gcBlobRuleCronRule == "" ? null : (
                                  <p className="mt-1 text-xs text-gray-600">
                                    <span>
                                      Next run at '{gcBlobRuleCronRuleNextRunAt}'.
                                    </span>
                                  </p>
                                )
                              }
                            </div>
                          </div>
                        </>
                      )
                    }
                    <div className="flex flex-row-reverse mt-5">
                      <button
                        type="button"
                        className="inline-flex w-full justify-center rounded-md border border-transparent bg-indigo-500 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:bg-indigo-500 focus:ring-offset-2 sm:ml-3 sm:w-auto sm:text-sm"
                        onClick={e => createOrUpdateGcBlob()}
                      >
                        {
                          gcBlobRuleExist ? "Update" : "Create"
                        }
                      </button>
                      <button
                        type="button"
                        className="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:mt-0 sm:w-auto sm:text-sm"
                        onClick={e => { setGcBlobRuleConfigModal(false) }}
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                </Dialog.Panel>
              </Transition.Child>
            </div>
          </div>
        </Dialog>
      </Transition.Root>
      <Transition.Root show={gcRepositoryRuleConfigModal} as={Fragment}>
        <Dialog as="div" className="relative z-10" onClose={setGcRepositoryRuleConfigModal}>
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
          </Transition.Child>

          <div className="fixed inset-0 z-10 overflow-y-auto">
            <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
              <Transition.Child
                as={Fragment}
                enter="ease-out duration-300"
                enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                enterTo="opacity-100 translate-y-0 sm:scale-100"
                leave="ease-in duration-200"
                leaveFrom="opacity-100 translate-y-0 sm:scale-100"
                leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              >
                <Dialog.Panel className="relative transform overflow-hidden rounded-lg bg-white px-6 pb-4 text-left shadow-xl transition-all">
                  <Dialog.Title
                    as="h3"
                    className="text-lg font-medium leading-6 text-gray-900 border-b pt-4 pb-4"
                  >
                    Garbage collect empty repository config
                  </Dialog.Title>
                  <div className="flex flex-col gap-0 mt-4">
                    <div className="grid grid-cols-6 gap-4">
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            <span className="text-red-600">*</span>
                            <span className="leading-6 ">Retention Days</span>
                            <div className="flex flex-row cursor-pointer"
                              id="gcRepositoryRetentionDaysHelp"
                              onClick={e => {
                                let tooltip = new Tooltip(document.getElementById("tooltip-gc-repository-retention-days"),
                                  document.getElementById("gcRepositoryRetentionDaysHelp"), { triggerType: "click" });
                                tooltip.show();
                              }}
                            >
                              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 block my-auto ml-0.5">
                                <path strokeLinecap="round" strokeLinejoin="round" d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z" />
                              </svg>
                            </div>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-4">
                        <div className="relative rounded-md shadow-sm">
                          <input
                            type="text"
                            id="namespace_count_limit"
                            name="namespace_count_limit"
                            placeholder="0 means no limit"
                            className={(gcRepositoryRuleRetentionDaysValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                            value={gcRepositoryRuleRetentionDays}
                            onChange={e => setGcRepositoryRuleRetentionDays(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))}
                          />
                          {
                            gcRepositoryRuleRetentionDaysValid ? null : (
                              <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-red-500">
                                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                                </svg>
                              </div>
                            )
                          }
                        </div>
                      </div>
                    </div>
                    <div className="grid grid-cols-6 gap-4">
                      <div className="col-span-2"></div>
                      <div className="col-span-4">
                        {
                          gcRepositoryRuleRetentionDaysValid ? null : (
                            <p className="mt-1 text-xs text-red-600">
                              <span>
                                Not a valid retention days limit, available 0-180.
                              </span>
                            </p>
                          )
                        }
                      </div>
                    </div>
                    <div className="grid grid-cols-6 gap-4 mt-4">
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            <span className="text-red-600">*</span>
                            <span className="leading-6 ">Cron Enabled</span>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-4">
                        <div className="mt-0.5 flex flex-row items-center h-[36px]">
                          <label className="relative inline-flex items-center cursor-pointer">
                            <input type="checkbox" checked={gcRepositoryRuleCronEnabled} className="sr-only peer" onChange={e => {
                              setGcRepositoryRuleCronEnabled(e.target.checked);
                            }} />
                            <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                          </label>
                        </div>
                      </div>
                    </div>
                    {
                      !gcRepositoryRuleCronEnabled ? null : (
                        <>
                          <div className="grid grid-cols-6 gap-4 mt-4">
                            <div className="col-span-2 flex flex-row">
                              <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                                <div className="flex">
                                  <span className="text-red-600">*</span>
                                  <span className="leading-6 ">Cron Rule</span>
                                  <div className="flex flex-row cursor-pointer"
                                    id="gcRepositoryRuleHelp"
                                    onClick={e => {
                                      let tooltip = new Tooltip(document.getElementById("tooltip-gc-repository-cron-rule"),
                                        document.getElementById("gcRepositoryRuleHelp"), { triggerType: "click" });
                                      tooltip.show();
                                    }}
                                  >
                                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 block my-auto ml-0.5">
                                      <path strokeLinecap="round" strokeLinejoin="round" d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z" />
                                    </svg>
                                  </div>
                                  <span>:</span>
                                </div>
                              </label>
                            </div>

                            <div className="col-span-4">
                              <div className="relative rounded-md shadow-sm">
                                <input
                                  type="text"
                                  id="gc_repository_cron_rule"
                                  name="gc_repository_cron_rule"
                                  placeholder="cron rule"
                                  className={(gcRepositoryRuleCronRuleValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                                  value={gcRepositoryRuleCronRule}
                                  onChange={e => setGcRepositoryRuleCronRule(e.target.value)}
                                />
                                {
                                  gcRepositoryRuleCronRuleValid ? null : (
                                    <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-red-500">
                                        <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                                      </svg>
                                    </div>
                                  )
                                }
                              </div>

                            </div>
                          </div>
                          <div className="grid grid-cols-6 gap-4">
                            <div className="col-span-2">
                            </div>
                            <div className="col-span-4">
                              {
                                !gcRepositoryRuleCronRuleValid ? (
                                  <p className="mt-1 text-xs text-red-600">
                                    <span>
                                      Not a valid cron rule, you can try '0 0 * * 6'.
                                    </span>
                                  </p>
                                ) : gcRepositoryRuleCronRule == "" ? null : (
                                  <p className="mt-1 text-xs text-gray-600">
                                    <span>
                                      Next run at '{gcRepositoryRuleCronRuleNextRunAt}'.
                                    </span>
                                  </p>
                                )
                              }
                            </div>
                          </div>
                        </>
                      )
                    }
                    <div className="flex flex-row-reverse mt-5">
                      <button
                        type="button"
                        className="inline-flex w-full justify-center rounded-md border border-transparent bg-indigo-500 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:bg-indigo-500 focus:ring-offset-2 sm:ml-3 sm:w-auto sm:text-sm"
                        onClick={e => createOrUpdateGcRepository()}
                      >
                        {
                          gcRepositoryRuleExist ? "Update" : "Create"
                        }
                      </button>
                      <button
                        type="button"
                        className="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:mt-0 sm:w-auto sm:text-sm"
                        onClick={e => { setGcRepositoryRuleConfigModal(false) }}
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                </Dialog.Panel>
              </Transition.Child>
            </div>
          </div>
        </Dialog>
      </Transition.Root>
      <Transition.Root show={gcArtifactRuleConfigModal} as={Fragment}>
        <Dialog as="div" className="relative z-10" onClose={setGcArtifactRuleConfigModal}>
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
          </Transition.Child>
          <div className="fixed inset-0 z-10 overflow-y-auto">
            <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
              <Transition.Child
                as={Fragment}
                enter="ease-out duration-300"
                enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                enterTo="opacity-100 translate-y-0 sm:scale-100"
                leave="ease-in duration-200"
                leaveFrom="opacity-100 translate-y-0 sm:scale-100"
                leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              >
                <Dialog.Panel className="relative transform overflow-hidden rounded-lg bg-white px-6 pb-4 text-left shadow-xl transition-all">
                  <Dialog.Title
                    as="h3"
                    className="text-lg font-medium leading-6 text-gray-900 border-b pt-4 pb-4"
                  >
                    Garbage collect artifact config
                  </Dialog.Title>
                  <div className="flex flex-col gap-0 mt-4">
                    <div className="grid grid-cols-6 gap-4">
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            <span className="text-red-600">*</span>
                            <span className="leading-6 ">Retention Days</span>
                            <div className="flex flex-row cursor-pointer"
                              id="gcArtifactRetentionDaysHelp"
                              onClick={e => {
                                let tooltip = new Tooltip(document.getElementById("tooltip-gc-artifact-retention-days"),
                                  document.getElementById("gcArtifactRetentionDaysHelp"), { triggerType: "click" });
                                tooltip.show();
                              }}
                            >
                              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 block my-auto ml-0.5">
                                <path strokeLinecap="round" strokeLinejoin="round" d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z" />
                              </svg>
                            </div>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-4">
                        <div className="relative rounded-md shadow-sm">
                          <input
                            type="text"
                            id="namespace_count_limit"
                            name="namespace_count_limit"
                            placeholder="0 means no limit"
                            className={(gcArtifactRuleRetentionDaysValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                            value={gcArtifactRuleRetentionDays}
                            onChange={e => setGcArtifactRuleRetentionDays(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))}
                          />
                          {
                            gcArtifactRuleRetentionDaysValid ? null : (
                              <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-red-500">
                                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                                </svg>
                              </div>
                            )
                          }
                        </div>
                      </div>
                    </div>
                    <div className="grid grid-cols-6 gap-4">
                      <div className="col-span-2"></div>
                      <div className="col-span-4">
                        {
                          gcArtifactRuleRetentionDaysValid ? null : (
                            <p className="mt-1 text-xs text-red-600">
                              <span>
                                Not a valid retention days limit, available 0-180.
                              </span>
                            </p>
                          )
                        }
                      </div>
                    </div>
                    <div className="grid grid-cols-6 gap-4 mt-4">
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            <span className="text-red-600">*</span>
                            <span className="leading-6 ">Cron Enabled</span>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-4">
                        <div className="mt-0.5 flex flex-row items-center h-[36px]">
                          <label className="relative inline-flex items-center cursor-pointer">
                            <input type="checkbox" checked={gcArtifactCronEnabled} className="sr-only peer" onChange={e => {
                              setGcArtifactCronEnabled(e.target.checked);
                            }} />
                            <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                          </label>
                        </div>
                      </div>
                    </div>
                    {
                      !gcArtifactCronEnabled ? null : (
                        <>
                          <div className="grid grid-cols-6 gap-4 mt-4">
                            <div className="col-span-2 flex flex-row">
                              <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                                <div className="flex">
                                  <span className="text-red-600">*</span>
                                  <span className="leading-6 ">Cron Rule</span>
                                  <div className="flex flex-row cursor-pointer"
                                    id="gcArtifactRuleHelp"
                                    onClick={e => {
                                      let tooltip = new Tooltip(document.getElementById("tooltip-gc-artifact-cron-rule"),
                                        document.getElementById("gcArtifactRuleHelp"), { triggerType: "click" });
                                      tooltip.show();
                                    }}
                                  >
                                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 block my-auto ml-0.5">
                                      <path strokeLinecap="round" strokeLinejoin="round" d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z" />
                                    </svg>
                                  </div>
                                  <span>:</span>
                                </div>
                              </label>
                            </div>
                            <div className="col-span-4">
                              <div className="relative rounded-md shadow-sm">
                                <input
                                  type="text"
                                  id="gc_repository_cron_rule"
                                  name="gc_repository_cron_rule"
                                  placeholder="cron rule"
                                  className={(gcArtifactRuleCronRuleValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                                  value={gcArtifactRuleCronRule}
                                  onChange={e => setGcArtifactRuleCronRule(e.target.value)}
                                />
                                {
                                  gcArtifactRuleCronRuleValid ? null : (
                                    <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-red-500">
                                        <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                                      </svg>
                                    </div>
                                  )
                                }
                              </div>
                            </div>
                          </div>
                          <div className="grid grid-cols-6 gap-4">
                            <div className="col-span-2">
                            </div>
                            <div className="col-span-4">
                              {
                                !gcArtifactRuleCronRuleValid ? (
                                  <p className="mt-1 text-xs text-red-600">
                                    <span>
                                      Not a valid cron rule, you can try '0 0 * * 6'.
                                    </span>
                                  </p>
                                ) : gcArtifactRuleCronRule == "" ? null : (
                                  <p className="mt-1 text-xs text-gray-600">
                                    <span>
                                      Next run at '{gcArtifactRuleCronRuleNextRunAt}'.
                                    </span>
                                  </p>
                                )
                              }
                            </div>
                          </div>
                        </>
                      )
                    }
                    <div className="flex flex-row-reverse mt-5">
                      <button
                        type="button"
                        className="inline-flex w-full justify-center rounded-md border border-transparent bg-indigo-500 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:bg-indigo-500 focus:ring-offset-2 sm:ml-3 sm:w-auto sm:text-sm"
                        onClick={e => createOrUpdateGcArtifact()}
                      >
                        {
                          gcArtifactRuleExist ? "Update" : "Create"
                        }
                      </button>
                      <button
                        type="button"
                        className="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:mt-0 sm:w-auto sm:text-sm"
                        onClick={e => { setGcArtifactRuleConfigModal(false) }}
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                </Dialog.Panel>
              </Transition.Child>
            </div>
          </div>
        </Dialog>
      </Transition.Root>
      <Transition.Root show={gcTagRuleConfigModal} as={Fragment}>
        <Dialog as="div" className="relative z-10" onClose={setGcTagRuleConfigModal}>
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
          </Transition.Child>
          <div className="fixed inset-0 z-10 overflow-y-auto">
            <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
              <Transition.Child
                as={Fragment}
                enter="ease-out duration-300"
                enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                enterTo="opacity-100 translate-y-0 sm:scale-100"
                leave="ease-in duration-200"
                leaveFrom="opacity-100 translate-y-0 sm:scale-100"
                leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              >
                <Dialog.Panel className="relative transform overflow-hidden rounded-lg bg-white px-6 pb-4 text-left shadow-xl transition-all min-w-[700px]">
                  <Dialog.Title
                    as="h3"
                    className="text-lg font-medium leading-6 text-gray-900 border-b pt-4 pb-4"
                  >
                    Garbage collect tag config
                  </Dialog.Title>
                  <div className="flex flex-col gap-0 mt-4">
                    <div className="grid grid-cols-8 gap-4">
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            <span className="text-red-600">*</span>
                            <span className="leading-6 ">Retention Amount</span>
                            <div className="flex flex-row cursor-pointer"
                              id="gcTagRetentionAmountHelp"
                              onClick={e => {
                                let tooltip = new Tooltip(document.getElementById("tooltip-gc-tag-retention-amount"),
                                  document.getElementById("gcTagRetentionAmountHelp"), { triggerType: "click" });
                                tooltip.show();
                              }}
                            >
                              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 block my-auto ml-0.5">
                                <path strokeLinecap="round" strokeLinejoin="round" d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z" />
                              </svg>
                            </div>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-2">
                        <Listbox
                          value={gcTagRuleRetentionType}
                          onChange={(source: string) => {
                            setGcTagRuleRetentionType(source);
                          }}>
                          <div className="relative w-full">
                            <Listbox.Button
                              className={() => {
                                let cursor = ''
                                if ((searchParams.get('code_repository_stick') || '') === 'true') {
                                  cursor = 'cursor-not-allowed ';
                                } else {
                                  cursor = 'cursor-pointer ';
                                }
                                return cursor + "relative w-full rounded-lg bg-white py-2 pl-3 pr-10 text-left shadow focus:outline-none focus-visible:border-indigo-500 focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-orange-300 sm:text-sm"
                              }}
                            >
                              <span className="block truncate">{gcTagRuleRetentionType}</span>
                              <span className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
                                <ChevronUpDownIcon
                                  className="h-5 w-5 text-gray-400"
                                  aria-hidden="true"
                                />
                              </span>
                            </Listbox.Button>
                            <Listbox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm z-10">
                              <Transition
                                leave="transition ease-in duration-100"
                                leaveFrom="opacity-100"
                                leaveTo="opacity-0"
                              >
                                {
                                  retentionAmountType.map(source => (
                                    <Listbox.Option key={source.name} value={source.name} className={({ active }) =>
                                      `relative cursor-pointer select-none py-2 pl-10 pr-4 ${active ? 'bg-gray-100 text-gray-900' : 'text-gray-900'
                                      }`
                                    }>
                                      {({ selected }) => (
                                        <>
                                          <span
                                            className={`block truncate ${selected ? 'font-medium' : 'font-normal'
                                              }`}
                                          >
                                            {source.name}
                                          </span>
                                          {
                                            selected ? (
                                              <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-gray-600">
                                                <CheckIcon className="h-5 w-5" aria-hidden="true" />
                                              </span>
                                            ) : null
                                          }
                                        </>
                                      )}
                                    </Listbox.Option>
                                  ))
                                }
                              </Transition>
                            </Listbox.Options>
                          </div>
                        </Listbox>
                      </div>
                      <div className="col-span-4">
                        <div className="relative rounded-md shadow-sm">
                          <input
                            type="text"
                            id="namespace_count_limit"
                            name="namespace_count_limit"
                            placeholder="0 means no limit"
                            className={(gcTagRuleRetentionAmountValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                            value={gcTagRuleRetentionAmount}
                            onChange={e => setGcTagRuleRetentionAmount(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))}
                          />
                          {
                            gcTagRuleRetentionAmountValid ? null : (
                              <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-red-500">
                                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                                </svg>
                              </div>
                            )
                          }
                        </div>
                      </div>
                    </div>
                    <div className="grid grid-cols-8 gap-4">
                      <div className="col-span-2"></div>
                      <div className="col-span-2"></div>
                      <div className="col-span-4">
                        {
                          gcTagRuleRetentionAmountValid ? null : (
                            <p className="mt-1 text-xs text-red-600">
                              <span>
                                Not a valid retention amount, available 1-180.
                              </span>
                            </p>
                          )
                        }
                      </div>
                    </div>
                    <div className="grid grid-cols-8 mt-4 gap-4">
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            {/* <span className="text-red-600">*</span> */}
                            <span className="leading-6 ">Retention Regex</span>
                            <div className="flex flex-row cursor-pointer"
                              id="gcTagRetentionPatternHelp"
                              onClick={e => {
                                let tooltip = new Tooltip(document.getElementById("tooltip-gc-tag-retention-pattern"),
                                  document.getElementById("gcTagRetentionPatternHelp"), { triggerType: "click" });
                                tooltip.show();
                              }}
                            >
                              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 block my-auto ml-0.5">
                                <path strokeLinecap="round" strokeLinejoin="round" d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z" />
                              </svg>
                            </div>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-6">
                        <div className="relative rounded-md shadow-sm">
                          <input
                            type="text"
                            id="namespace_count_limit"
                            name="namespace_count_limit"
                            placeholder="regexp"
                            className={(gcTagRuleRetentionPatternValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                            value={gcTagRuleRetentionPattern}
                            onChange={e => setGcTagRuleRetentionPattern(e.target.value)}
                          />
                          {
                            gcTagRuleRetentionPatternValid ? null : (
                              <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-red-500">
                                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                                </svg>
                              </div>
                            )
                          }
                        </div>
                      </div>
                    </div>
                    <div className="grid grid-cols-8 gap-4">
                      <div className="col-span-2"></div>
                      <div className="col-span-6">
                        {
                          gcTagRuleRetentionPatternValid ? null : (
                            <p className="mt-1 text-xs text-red-600">
                              <span>
                                Not a valid regex, you can try 'v.*'.
                              </span>
                            </p>
                          )
                        }
                      </div>
                    </div>
                    <div className="grid grid-cols-8 mt-4 gap-4">
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            <span className="text-red-600">*</span>
                            <span className="leading-6 ">Cron Enabled</span>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-6">
                        <div className="mt-0.5 flex flex-row items-center h-[36px]">
                          <label className="relative inline-flex items-center cursor-pointer">
                            <input type="checkbox" checked={gcTagRuleCronEnabled} className="sr-only peer" onChange={e => {
                              setGcTagRuleCronEnabled(e.target.checked);
                            }} />
                            <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                          </label>
                        </div>
                      </div>
                    </div>
                    {
                      !gcTagRuleCronEnabled ? null : (
                        <>
                          <div className="grid grid-cols-8 gap-4 mt-4">
                            <div className="col-span-2 flex flex-row">
                              <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                                <div className="flex">
                                  <span className="text-red-600">*</span>
                                  <span className="leading-6 ">Cron Rule</span>
                                  <div className="flex flex-row cursor-pointer"
                                    id="gcTagRuleHelp"
                                    onClick={e => {
                                      let tooltip = new Tooltip(document.getElementById("tooltip-gc-tag-cron-rule"),
                                        document.getElementById("gcTagRuleHelp"), { triggerType: "click" });
                                      tooltip.show();
                                    }}
                                  >
                                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 block my-auto ml-0.5">
                                      <path strokeLinecap="round" strokeLinejoin="round" d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z" />
                                    </svg>
                                  </div>
                                  <span>:</span>
                                </div>
                              </label>
                            </div>
                            <div className="col-span-6">
                              <div className="relative rounded-md shadow-sm">
                                <input
                                  type="text"
                                  id="gc_repository_cron_rule"
                                  name="gc_repository_cron_rule"
                                  placeholder="cron rule"
                                  className={(gcTagRuleCronRuleValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                                  value={gcTagRuleCronRule}
                                  onChange={e => setGcTagRuleCronRule(e.target.value)}
                                />
                                {
                                  gcTagRuleCronRuleValid ? null : (
                                    <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-red-500">
                                        <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                                      </svg>
                                    </div>
                                  )
                                }
                              </div>
                            </div>
                          </div>
                          <div className="grid grid-cols-8 gap-4">
                            <div className="col-span-2">
                            </div>
                            <div className="col-span-6">
                              {
                                !gcTagRuleCronRuleValid ? (
                                  <p className="mt-1 text-xs text-red-600">
                                    <span>
                                      Not a valid cron rule, you can try '0 0 * * 6'.
                                    </span>
                                  </p>
                                ) : gcTagRuleCronRule == "" ? null : (
                                  <p className="mt-1 text-xs text-gray-600">
                                    <span>
                                      Next run at '{gcTagRuleCronRuleNextRunAt}'.
                                    </span>
                                  </p>
                                )
                              }
                            </div>
                          </div>
                        </>
                      )
                    }
                    <div className="flex flex-row-reverse mt-5">
                      <button
                        type="button"
                        className="inline-flex w-full justify-center rounded-md border border-transparent bg-indigo-500 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:bg-indigo-500 focus:ring-offset-2 sm:ml-3 sm:w-auto sm:text-sm"
                        onClick={e => createOrUpdateGcTag()}
                      >
                        {
                          gcTagRuleExist ? "Update" : "Create"
                        }
                      </button>
                      <button
                        type="button"
                        className="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:mt-0 sm:w-auto sm:text-sm"
                        onClick={e => { setGcTagRuleConfigModal(false) }}
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                </Dialog.Panel>
              </Transition.Child>
            </div>
          </div>
        </Dialog>
      </Transition.Root>
    </Fragment >
  )
}
