/*
J13,2  SPI1_SCLK  SOM-167   TS (GPIO3_0)

J13,3  I2C1_SDA  SOM-104  SDA
J13,5  I2C1_SCL  SOM-106  SCL
*/

#define GESTIC_GPIO_TS ((3*32) + 10) // 3_0 on EVK
#define GESTIC_GPIO_TS_NAME "mii1_rxclk" // mii1_col on EVK

#define GESTIC_GPIO_MCLR ((3*32) + 4) // 3_0 on EVK
#define GESTIC_GPIO_MCLR_NAME "mii1_rxdv"

#define GESTIC_I2C_BUS_NUM 2
#define GESTIC_I2C_ADDRESS 0x42


#define GESTIC_DEBUG 0


#include <linux/module.h>
#include <linux/i2c.h>
#include <linux/version.h>
#include <linux/kernel.h>
#include <linux/types.h>
#include <linux/kdev_t.h>
#include <linux/fs.h>
#include <linux/device.h>
#include <linux/cdev.h>
#include <linux/poll.h>
#include <linux/wait.h>
#include <linux/gpio.h>
#include <linux/irq.h>
#include <linux/interrupt.h>
#include <../arch/arm/mach-omap2/mux.h>

static dev_t first_dev;
static struct cdev c_dev;
static struct class *cls;
static int ts_irq_num;
static wait_queue_head_t read_queue;
static struct i2c_client *gestic_client;


struct gestic_message_header
{
  uint8_t size;
  uint8_t flags;
  uint8_t seq;
  uint8_t id;
} __attribute__( ( packed ) );


static struct i2c_device_id gestic_idtable[] = {
  { "gestic", GESTIC_I2C_ADDRESS },
  { }
};

MODULE_DEVICE_TABLE(i2c, gestic_idtable);

static struct i2c_board_info gestic_boardinfo = {
  I2C_BOARD_INFO("gestic", 0x42)
};


static int gestic_open(struct inode *i, struct file *f)
{
  if (GESTIC_DEBUG) printk(KERN_INFO "GestIC: open()\n");
  return 0;
}

static int gestic_close(struct inode *i, struct file *f)
{
  if (GESTIC_DEBUG) printk(KERN_INFO "GestIC: close()\n");
  return 0;
}

static ssize_t gestic_read(struct file *f, char __user *buf, size_t
  len, loff_t *off)
{
  char ktmp[512];
  struct gestic_message_header *hdr = (struct gestic_message_header *)ktmp;
  int bytes_read = 0;

  if (GESTIC_DEBUG) printk(KERN_INFO "GestIC: read()\n");

  // make sure the TS line is asserted
  if (gpio_get_value(GESTIC_GPIO_TS) == 0)
  {
    // asserted: now we assert too, so the data doesn't change
    gpio_direction_output(GESTIC_GPIO_TS, 0);
  }
  else
  {
    // not asserted: failed to read
    return 0;
  }

  // perform the i2c transactions
  if (GESTIC_DEBUG) printk(KERN_INFO "GestIC: ready to read (client=%p, addr=%d)\n", gestic_client, gestic_client->addr);

  bytes_read = i2c_master_recv(gestic_client, ktmp, 255);

  if (bytes_read > 0 ) {
    printk(KERN_INFO "GestIC: read %d bytes\n", bytes_read);
    printk(KERN_INFO "GestIC: msg[size=%d, flags=%d, seq=%d, id=%d]\n", hdr->size, hdr->flags, hdr->seq, hdr->id);

    // cap the bytes to the amount in the actual GestIC payload
    if ( bytes_read > hdr->size )
      bytes_read = hdr->size;

    // cap the bytes at what the user requested
    if ( bytes_read > len )
      bytes_read = len;

    // copy bytes, noting for the user how many actually arrived (copy_to_user returns bytes NOT copied)
    bytes_read -= copy_to_user(buf, ktmp, bytes_read);
  } else {
    if (GESTIC_DEBUG) printk(KERN_INFO "GestIC: read FAILED with error=%d\n", -bytes_read);
    bytes_read = 0;
  }

  // stop asserting the TS line
  gpio_direction_input(GESTIC_GPIO_TS);

  return bytes_read;
}

static ssize_t gestic_write(struct file *f, const char __user *buf,
  size_t len, loff_t *off)
{
  char ktmp[512];
  struct gestic_message_header *hdr = (struct gestic_message_header *)ktmp;
  size_t bytes = len;
  int bytes_sent;

  if (GESTIC_DEBUG) printk(KERN_INFO "GestIC: write( %d bytes )\n", len);
  
  if ( bytes > 255 ) {
    bytes = 255;
  }

  bytes -= copy_from_user(ktmp, buf, bytes);

  bytes_sent = i2c_master_send(gestic_client, ktmp, bytes);

  if ( bytes_sent < 0 ) {
    bytes_sent = 0;
  }

  if (GESTIC_DEBUG) printk(KERN_INFO "GestIC: msg[size=%d, flags=%d, seq=%d, id=%d]\n", hdr->size, hdr->flags, hdr->seq, hdr->id);
  if (GESTIC_DEBUG) printk(KERN_INFO "GestIC: bytes in=%d, req=%d, sent=%d\n", len, bytes, bytes_sent);

  return bytes_sent;
}

static unsigned int gestic_poll(struct file *f, poll_table *pt)
{
  unsigned int mask = 0;

  if (GESTIC_DEBUG) printk(KERN_INFO "GestIC: polled");

  // GESTIC_GPIO_TS pulled low (asserted) when data available for reading
  if (gpio_get_value(GESTIC_GPIO_TS) == 0) mask |= POLLIN | POLLRDNORM;
  //if (data_avail_to_write) mask |= POLLOUT | POLLWRNORM;

  poll_wait(f, &read_queue, pt);
  //poll_wait(file, &write_queue, pt);

  return mask;
}

static struct file_operations pugs_fops =
{
  .owner = THIS_MODULE,
  .open = gestic_open,
  .release = gestic_close,
  .read = gestic_read,
  .write = gestic_write,
  .poll = gestic_poll
};

static irqreturn_t data_incoming_ready(int irq, void *dev_id)
{
  if ( irq != ts_irq_num )
  {
    return IRQ_NONE;
  }

  if (GESTIC_DEBUG) printk(KERN_INFO "GestIC: TS asserted");

  wake_up(&read_queue);

  return IRQ_HANDLED;
}

static int __init gestic_init(void)
{
  int err;
  struct i2c_adapter *i2c_adap;

  printk(KERN_INFO "GestIC: registered");

  init_waitqueue_head(&read_queue);

  if (alloc_chrdev_region(&first_dev, 0, 1, "GestIC") < 0)
  {
    return -1;
  }

  if ((cls = class_create(THIS_MODULE, "chardrv")) == NULL)
  {
    unregister_chrdev_region(first_dev, 1);
    return -1;
  }

  if (device_create(cls, NULL, first_dev, NULL, "gestic") == NULL)
  {
    class_destroy(cls);
    unregister_chrdev_region(first_dev, 1);
    return -1;
  }

  cdev_init(&c_dev, &pugs_fops);
  if (cdev_add(&c_dev, first_dev, 1) == -1)
  {
    device_destroy(cls, first_dev);
    class_destroy(cls);
    unregister_chrdev_region(first_dev, 1);
    return -1;
  }

  i2c_adap = i2c_get_adapter(GESTIC_I2C_BUS_NUM);
  gestic_client = i2c_new_device(i2c_adap, &gestic_boardinfo);

  // prep the GPIO for the TS line
  //omap_mux_init_gpio(GESTIC_GPIO_TS, OMAP_PIN_INPUT);
  err = gpio_request(GESTIC_GPIO_TS, "mgc3130_TS");
  gpio_direction_input(GESTIC_GPIO_TS);
  ts_irq_num = gpio_to_irq(GESTIC_GPIO_TS);
  if (err == -1 || ts_irq_num == -1)
  {
    i2c_unregister_device(gestic_client);
    cdev_del(&c_dev);
    device_destroy(cls, first_dev);
    class_destroy(cls);
    unregister_chrdev_region(first_dev, 1);
    return -1;
  }

  if (request_irq(ts_irq_num, data_incoming_ready, 0, "mgc3130_TS_R", NULL) == -1)
  {
    gpio_free(GESTIC_GPIO_TS);
    i2c_unregister_device(gestic_client);
    cdev_del(&c_dev);
    device_destroy(cls, first_dev);
    class_destroy(cls);
    unregister_chrdev_region(first_dev, 1);
    return -1;
  }

  irq_set_irq_type(ts_irq_num, IRQ_TYPE_EDGE_FALLING);

  return 0;
}
 
static void __exit gestic_exit(void)
{
  free_irq(ts_irq_num, NULL);
  gpio_free(GESTIC_GPIO_TS);

  i2c_unregister_device(gestic_client);

  cdev_del(&c_dev);
  device_destroy(cls, first_dev);
  class_destroy(cls);
  unregister_chrdev_region(first_dev, 1);

  printk(KERN_INFO "GestIC: unregistered");
}
 
module_init(gestic_init);
module_exit(gestic_exit);
MODULE_LICENSE("GPL");
MODULE_AUTHOR("Theo Julienne <theo@ninjablocks.com>");
MODULE_DESCRIPTION("MGC3130 GestIC Driver");
