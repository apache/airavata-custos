#!/usr/bin/python

# (c) 2020, NetApp, Inc
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type


DOCUMENTATION = """
---
module: na_santricity_iscsi_interface
short_description: NetApp E-Series manage iSCSI interface configuration
description:
    - Configure settings of an E-Series iSCSI interface
author:
    - Michael Price (@lmprice)
    - Nathan Swartz (@ndswartz)
extends_documentation_fragment:
    - netapp_eseries.santricity.santricity.santricity_doc
options:
    controller:
        description:
            - The controller that owns the port you want to configure.
            - Controller names are presented alphabetically, with the first controller as A,
             the second as B, and so on.
            - Current hardware models have either 1 or 2 available controllers, but that is not a guaranteed hard
             limitation and could change in the future.
        type: str
        required: true
        choices:
            - A
            - B
    port:
        description:
            - The controller iSCSI HIC port to modify.
            - You can determine this value by numbering the iSCSI ports left to right on the controller you wish to modify starting with one.
        type: int
        required: true
    state:
        description:
            - When enabled, the provided configuration will be utilized.
            - When disabled, the IPv4 configuration will be cleared and IPv4 connectivity disabled.
        type: str
        choices:
            - enabled
            - disabled
        default: enabled
    address:
        description:
            - The IPv4 address to assign to the interface.
            - Should be specified in xx.xx.xx.xx form.
            - Mutually exclusive with I(config_method=dhcp)
        type: str
        required: false
    subnet_mask:
        description:
            - The subnet mask to utilize for the interface.
            - Should be specified in xx.xx.xx.xx form.
            - Mutually exclusive with I(config_method=dhcp)
        type: str
    gateway:
        description:
            - The IPv4 gateway address to utilize for the interface.
            - Should be specified in xx.xx.xx.xx form.
            - Mutually exclusive with I(config_method=dhcp)
        type: str
        required: false
    config_method:
        description:
            - The configuration method type to use for this interface.
            - dhcp is mutually exclusive with I(address), I(subnet_mask), and I(gateway).
        type: str
        choices:
            - dhcp
            - static
        default: dhcp
        required: false
    mtu:
        description:
            - The maximum transmission units (MTU), in bytes.
            - This allows you to configure a larger value for the MTU, in order to enable jumbo frames
              (any value > 1500).
            - Generally, it is necessary to have your host, switches, and other components not only support jumbo
              frames, but also have it configured properly. Therefore, unless you know what you're doing, it's best to
              leave this at the default.
        type: int
        default: 1500
        required: false
        aliases:
            - max_frame_size
notes:
    - Check mode is supported.
    - The interface settings are applied synchronously, but changes to the interface itself (receiving a new IP address
      via dhcp, etc), can take seconds or minutes longer to take effect.
    - This module will not be useful/usable on an E-Series system without any iSCSI interfaces.
    - This module requires a Web Services API version of >= 1.3.
"""

EXAMPLES = """
    - name: Configure the first port on the A controller with a static IPv4 address
      na_santricity_iscsi_interface:
        ssid: "1"
        api_url: "https://192.168.1.100:8443/devmgr/v2"
        api_username: "admin"
        api_password: "adminpass"
        validate_certs: true
        port: "1"
        controller: "A"
        config_method: static
        address: "192.168.1.100"
        subnet_mask: "255.255.255.0"
        gateway: "192.168.1.1"

    - name: Disable ipv4 connectivity for the second port on the B controller
      na_santricity_iscsi_interface:
        ssid: "1"
        api_url: "https://192.168.1.100:8443/devmgr/v2"
        api_username: "admin"
        api_password: "adminpass"
        validate_certs: true
        port: "2"
        controller: "B"
        state: disabled

    - name: Enable jumbo frames for the first 4 ports on controller A
      na_santricity_iscsi_interface:
        ssid: "1"
        api_url: "https://192.168.1.100:8443/devmgr/v2"
        api_username: "admin"
        api_password: "adminpass"
        validate_certs: true
        port: "{{ item }}"
        controller: "A"
        state: enabled
        mtu: 9000
        config_method: dhcp
      loop:
        - 1
        - 2
        - 3
        - 4
"""

RETURN = """
msg:
    description: Success message
    returned: on success
    type: str
    sample: The interface settings have been updated.
"""
import re

from ansible_collections.netapp_eseries.santricity.plugins.module_utils.santricity import NetAppESeriesModule
from ansible.module_utils._text import to_native


class NetAppESeriesIscsiInterface(NetAppESeriesModule):
    def __init__(self):
        ansible_options = dict(controller=dict(type="str", required=True, choices=["A", "B"]),
                               port=dict(type="int", required=True),
                               state=dict(type="str", required=False, default="enabled", choices=["enabled", "disabled"]),
                               address=dict(type="str", required=False),
                               subnet_mask=dict(type="str", required=False),
                               gateway=dict(type="str", required=False),
                               config_method=dict(type="str", required=False, default="dhcp", choices=["dhcp", "static"]),
                               mtu=dict(type="int", default=1500, required=False, aliases=["max_frame_size"]))

        required_if = [["config_method", "static", ["address", "subnet_mask"]]]
        super(NetAppESeriesIscsiInterface, self).__init__(ansible_options=ansible_options,
                                                          web_services_version="02.00.0000.0000",
                                                          required_if=required_if,
                                                          supports_check_mode=True)

        args = self.module.params
        self.controller = args["controller"]
        self.port = args["port"]
        self.mtu = args["mtu"]
        self.state = args["state"]
        self.address = args["address"]
        self.subnet_mask = args["subnet_mask"]
        self.gateway = args["gateway"]
        self.config_method = args["config_method"]

        self.check_mode = self.module.check_mode
        self.post_body = dict()
        self.controllers = list()
        self.get_target_interface_cache = None

        if self.mtu < 1500 or self.mtu > 9000:
            self.module.fail_json(msg="The provided mtu is invalid, it must be > 1500 and < 9000 bytes.")

        if self.config_method == "dhcp" and any([self.address, self.subnet_mask, self.gateway]):
            self.module.fail_json(msg="A config_method of dhcp is mutually exclusive with the address,"
                                      " subnet_mask, and gateway options.")

        # A relatively primitive regex to validate that the input is formatted like a valid ip address
        address_regex = re.compile(r"^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$")

        if self.address and not address_regex.match(self.address):
            self.module.fail_json(msg="An invalid ip address was provided for address.")

        if self.subnet_mask and not address_regex.match(self.subnet_mask):
            self.module.fail_json(msg="An invalid ip address was provided for subnet_mask.")

        if self.gateway and not address_regex.match(self.gateway):
            self.module.fail_json(msg="An invalid ip address was provided for gateway.")

    @property
    def interfaces(self):
        ifaces = list()
        try:
            rc, ifaces = self.request("storage-systems/%s/graph/xpath-filter?query=/controller/hostInterfaces" % self.ssid)
        except Exception as err:
            self.module.fail_json(msg="Failed to retrieve defined host interfaces. Array Id [%s]. Error [%s]." % (self.ssid, to_native(err)))

        # Filter out non-iSCSI interfaces
        iscsi_interfaces = []
        for iface in [iface for iface in ifaces if iface["interfaceType"] == "iscsi"]:
            if iface["iscsi"]["interfaceData"]["type"] == "ethernet":
                iscsi_interfaces.append(iface)

        return iscsi_interfaces

    def get_controllers(self):
        """Retrieve a mapping of controller labels to their references
        {
            "A": "070000000000000000000001",
            "B": "070000000000000000000002",
        }
        :return: the controllers defined on the system
        """
        controllers = list()
        try:
            rc, controllers = self.request("storage-systems/%s/graph/xpath-filter?query=/controller/id" % self.ssid)
        except Exception as err:
            self.module.fail_json(msg="Failed to retrieve controller list! Array Id [%s]. Error [%s]." % (self.ssid, to_native(err)))

        controllers.sort()

        controllers_dict = {}
        i = ord("A")
        for controller in controllers:
            label = chr(i)
            controllers_dict[label] = controller
            i += 1

        return controllers_dict

    def get_target_interface(self):
        """Retrieve the specific controller iSCSI interface."""
        if self.get_target_interface_cache is None:
            ifaces = self.interfaces

            controller_ifaces = []
            for iface in ifaces:
                if self.controllers[self.controller] == iface["iscsi"]["controllerId"]:
                    controller_ifaces.append([iface["iscsi"]["channel"], iface, iface["iscsi"]["interfaceData"]["ethernetData"]["linkStatus"]])

            sorted_controller_ifaces = sorted(controller_ifaces)
            if self.port < 1 or self.port > len(controller_ifaces):
                status_msg = ", ".join(["%s (link %s)" % (index + 1, values[2]) for index, values in enumerate(sorted_controller_ifaces)])
                self.module.fail_json(msg="Invalid controller %s iSCSI port. Available ports: %s, Array Id [%s]."
                                          % (self.controller, status_msg, self.ssid))

            self.get_target_interface_cache = sorted_controller_ifaces[self.port - 1][1]
        return self.get_target_interface_cache

    def make_update_body(self, target_iface):
        target_iface = target_iface["iscsi"]
        body = dict(iscsiInterface=target_iface["id"])
        update_required = False

        if self.state == "enabled":
            settings = dict()
            if not target_iface["ipv4Enabled"]:
                update_required = True
                settings["ipv4Enabled"] = [True]
            if self.mtu != target_iface["interfaceData"]["ethernetData"]["maximumFramePayloadSize"]:
                update_required = True
                settings["maximumFramePayloadSize"] = [self.mtu]
            if self.config_method == "static":
                ipv4Data = target_iface["ipv4Data"]["ipv4AddressData"]

                if ipv4Data["ipv4Address"] != self.address:
                    update_required = True
                    settings["ipv4Address"] = [self.address]
                if ipv4Data["ipv4SubnetMask"] != self.subnet_mask:
                    update_required = True
                    settings["ipv4SubnetMask"] = [self.subnet_mask]
                if self.gateway is not None and ipv4Data["ipv4GatewayAddress"] != self.gateway:
                    update_required = True
                    settings["ipv4GatewayAddress"] = [self.gateway]

                if target_iface["ipv4Data"]["ipv4AddressConfigMethod"] != "configStatic":
                    update_required = True
                    settings["ipv4AddressConfigMethod"] = ["configStatic"]

            elif target_iface["ipv4Data"]["ipv4AddressConfigMethod"] != "configDhcp":
                update_required = True
                settings.update(dict(ipv4Enabled=[True],
                                     ipv4AddressConfigMethod=["configDhcp"]))
            body["settings"] = settings

        else:
            if target_iface["ipv4Enabled"]:
                update_required = True
                body["settings"] = dict(ipv4Enabled=[False])

        return update_required, body

    def update(self):
        self.controllers = self.get_controllers()
        if self.controller not in self.controllers:
            self.module.fail_json(msg="The provided controller name is invalid. Valid controllers: %s." % ", ".join(self.controllers.keys()))

        iface_before = self.get_target_interface()
        update_required, body = self.make_update_body(iface_before)
        if update_required and not self.check_mode:
            try:
                rc, result = self.request("storage-systems/%s/symbol/setIscsiInterfaceProperties" % self.ssid, method="POST", data=body, ignore_errors=True)
                # We could potentially retry this a few times, but it's probably a rare enough case (unless a playbook
                #  is cancelled mid-flight), that it isn't worth the complexity.
                if rc == 422 and result["retcode"] in ["busy", "3"]:
                    self.module.fail_json(msg="The interface is currently busy (probably processing a previously requested modification request)."
                                              " This operation cannot currently be completed. Array Id [%s]. Error [%s]." % (self.ssid, result))
                # Handle authentication issues, etc.
                elif rc != 200:
                    self.module.fail_json(msg="Failed to modify the interface! Array Id [%s]. Error [%s]." % (self.ssid, to_native(result)))
            # This is going to catch cases like a connection failure
            except Exception as err:
                self.module.fail_json(msg="Connection failure: we failed to modify the interface! Array Id [%s]. Error [%s]." % (self.ssid, to_native(err)))

        self.module.exit_json(msg="The interface settings have been updated.", changed=update_required)


def main():
    iface = NetAppESeriesIscsiInterface()
    iface.update()


if __name__ == "__main__":
    main()
