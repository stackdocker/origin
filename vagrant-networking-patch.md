/opt/vagrant/embedded/gems/gems/vagrant-1.8.1/plugins/guests/fedora/cap/configure_networks.rb

          if virtual
            # machine.communicate.sudo("ls /sys/class/net | egrep -v lo\\|docker") do |_, result|
            machine.communicate.sudo("find -L /sys/class/net -xtype l -name device -maxdepth 2 2> /dev/null | awk -F/ '{print \$5}'") do |_, result|
              interface_names = result.split("\n")
            end

            interface_names_by_slot = networks.map do |network|
               "#{interface_names[network[:interface]]}"
            end

