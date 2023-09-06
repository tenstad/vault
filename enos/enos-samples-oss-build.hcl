# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1

sample "build_oss_linux_amd64_deb" {
  attributes = global.sample_attributes

  subset "smoke" {
    matrix {
      arch            = ["amd64"]
      artifact_source = ["crt"]
      artifact_type   = ["package"]
      distro          = ["ubuntu"]
      edition         = ["oss"]
    }
  }

  subset "upgrade" {
    matrix {
      arch            = ["amd64"]
      artifact_source = ["crt"]
      artifact_type   = ["package"]
      distro          = ["ubuntu"]
      edition         = ["oss"]
    }
  }
}

sample "build_oss_linux_arm64_deb" {
  attributes = global.sample_attributes

  subset "smoke" {
    matrix {
      arch            = ["arm64"]
      artifact_source = ["crt"]
      artifact_type   = ["package"]
      distro          = ["ubuntu"]
      edition         = ["oss"]
    }
  }

  subset "upgrade" {
    matrix {
      arch            = ["arm64"]
      artifact_source = ["crt"]
      artifact_type   = ["package"]
      distro          = ["ubuntu"]
      edition         = ["oss"]
    }
  }
}

sample "build_oss_linux_arm64_rpm" {
  attributes = global.sample_attributes

  subset "smoke" {
    matrix {
      arch            = ["arm64"]
      artifact_source = ["crt"]
      artifact_type   = ["package"]
      distro          = ["rhel"]
      edition         = ["oss"]
    }
  }

  subset "upgrade" {
    matrix {
      arch            = ["arm64"]
      artifact_source = ["crt"]
      artifact_type   = ["package"]
      distro          = ["rhel"]
      edition         = ["oss"]
    }
  }
}

sample "build_oss_linux_amd64_rpm" {
  attributes = global.sample_attributes

  subset "smoke" {
    matrix {
      arch            = ["amd64"]
      artifact_source = ["crt"]
      artifact_type   = ["package"]
      distro          = ["rhel"]
      edition         = ["oss"]
    }
  }

  subset "upgrade" {
    matrix {
      arch            = ["amd64"]
      artifact_source = ["crt"]
      artifact_type   = ["package"]
      distro          = ["rhel"]
      edition         = ["oss"]
    }
  }
}

sample "build_oss_linux_amd64_zip" {
  attributes = global.sample_attributes

  subset "smoke" {
    matrix {
      arch            = ["amd64"]
      artifact_type   = ["bundle"]
      artifact_source = ["crt"]
      edition         = ["oss"]
    }
  }

  subset "upgrade" {
    matrix {
      arch            = ["amd64"]
      artifact_type   = ["bundle"]
      artifact_source = ["crt"]
      edition         = ["oss"]
    }
  }
}

sample "build_oss_linux_arm64_zip" {
  attributes = global.sample_attributes

  subset "smoke" {
    matrix {
      arch            = ["arm64"]
      artifact_source = ["crt"]
      artifact_type   = ["bundle"]
      edition         = ["oss"]
    }
  }

  subset "upgrade" {
    matrix {
      arch            = ["arm64"]
      artifact_source = ["crt"]
      artifact_type   = ["bundle"]
      edition         = ["oss"]
    }
  }
}
