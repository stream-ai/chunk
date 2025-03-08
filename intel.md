# Developer Log - Intel NPU DX Project

## Watching "AI Vision Applications with OpenVINO" Part 2

**Time Spent**: 2 hours
**Task/Attempt**: follow video, install packages & succesfully start jupyter notebook
**Outcome**:
- jupyter notebook running cleanly on laptop in DevContainber environment
- DevContainer fully configured & repeatable, code checked in

**Issues/Friction Points**:
- There are lots of OpenVINO videos on YT. I'd like to see better organization of video playlists, more targeted to specific learning tasks.
- In video 2 (installation), the walkthrough suggests adding PyTorch/TF support to the OpenVINO download, but the actual install webpage doesn't support that. Video may be out of date.
- Video is geared to a Windows environment. However sometimes WSL brings it's own complexities. I'd like to see notes for WSL.
- Video's install instructions are very out of date, needed to use instructions at https://docs.openvino.ai/2025/get-started/install-openvino/install-openvino-pip.html
- It's pretty unclear if I need to install NPU drivers and, if so, exactly how to do thAT. I founbd https://github.com/intel/linux-npu-driver, but the instructions are quite sparse and assume a lot of familiarity with Intel products and frameworks (such as "Level Zero" etc)
- Lots of difficulty running the jupytrer lab due to missing dependencies, etc. Had to do a loyt of searching and researching which packages to install, stackoverflow, etc. Finally determined the correct libraries needed:
```bash
# ------------------------------------------------------
# Setup OpenVINO
# ------------------------------------------------------
RUN apt update && \
  apt install libgl1 \
    libgl1-mesa-glx \ 
    libglib2.0-0 -y

RUN pip install --upgrade pip && \
  pip install openvino && \
  pip install opencv-python && \
  pip install jupyter lab
```
(took about 2 hours to resolve)
- Jupyter lab gets a deprecation warning: `The `openvino.runtime` module is deprecated and will be removed in the 2026.0 release. Please replace `openvino.runtime` with `openvino`.`  Sinple fix, but just illustrates this codebase is not maintained.

**Notes**: