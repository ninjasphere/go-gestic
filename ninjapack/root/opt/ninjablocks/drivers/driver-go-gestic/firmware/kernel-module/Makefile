ARCH=arm
KDIR=~/variscite/linux-3.2.0-sdk06.00.00.00.var_r13

BUILD_VARS=ARCH=$(ARCH) CROSS_COMPILE=/usr/local/ti-sdk-am335x-evm/linux-devkit/sysroots/i686-arago-linux/usr/bin/arm-linux-gnueabihf-

obj-m += gestic.o

all:
	$(BUILD_VARS) make -C $(KDIR) M=$(PWD) modules

clean:
	$(BUILD_VARS) make -C $(KDIR) M=$(PWD) clean
