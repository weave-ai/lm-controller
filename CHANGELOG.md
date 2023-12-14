# Changelog

All notable changes to this project are documented in this file.

## 0.10.0

**Release date:** 2023-12-14

Improvements:
- Use llama-cpp-python v0.2.23 to support Mixtral 8x7B

## 0.9.0

**Release date:** 2023-12-12

Improvements:
- Use llama-cpp-python v0.2.22
- Expose the model's leve (base, instruct, chat) and stop words as the model metadata

## 0.8.1

**Release date:** 2023-11-29

Improvements:
- Use llama-cpp-python v0.2.20

## 0.8.0

**Release date:** 2023-11-27

Improvements:
- Add service type to the engine spec

## 0.7.0

**Release date:** 2023-11-27

Improvements:
- Add imagePullPolicy to the engine spec

## 0.6.0

**Release date:** 2023-11-27

Improvements:
- Implement model metadata 
- Implement resource estimation
- Update Llama-cpp-python to v0.2.19
Fixes:
- Default to rolling update when deployment does not have a PVC attached

## 0.5.0

**Release date:** 2023-11-24

Fixes:
- Fix number of threads to be core - 1 [PR #10](https://github.com/weave-ai/lm-controller/pull/10)

## 0.4.0

**Release date:** 2023-11-23

Fixes:
- Fix init container resource requests and limits. [PR #6](https://github.com/weave-ai/lm-controller/pull/6)
- Fix network-policy instruction in README

## 0.3.1

**Release date:** 2023-11-22

Fixes:
- Fix image name in the deployment YAML. [PR #3](https://github.com/weave-ai/lm-controller/pull/3)

## 0.3.0

**Release date:** 2023-11-22

Fixes:
- Fix RBAC namespace in the rbac.yaml file. [PR #2](https://github.com/weave-ai/lm-controller/pull/2)
- Improve the readme with installation instructions. [PR #2](https://github.com/weave-ai/lm-controller/pull/2)

## 0.2.1

**Release date:** 2023-11-22

Fixes:
- Add changelog file to the repo. [PR #1](https://github.com/weave-ai/lm-controller/pull/1)
- Add install instructions to the readme.
- Correct the version in the readme and the deployment YAML.

## 0.2.0

**Release date:** 2023-11-22

Fixes:
- Add missing RBAC yaml to the release page. 

## 0.1.0

**Release date:** 2023-11-22

This is the initial release of the LM-Controller.

Improvements:
- Initial release of the LM-Controller
- Release Pipeline
