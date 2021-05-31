package org.apache.custos.ssl.certificate.manager;

import org.quartz.CronScheduleBuilder;
import org.quartz.JobBuilder;
import org.quartz.JobDetail;
import org.quartz.Scheduler;
import org.quartz.SchedulerException;
import org.quartz.Trigger;
import org.quartz.TriggerBuilder;
import org.quartz.impl.StdSchedulerFactory;

public class Launcher {
    public static void main(String[] args) {
        JobDetail job = JobBuilder
                .newJob(CertUpdater.class)
                .withIdentity("cert-updater")
                .build();
        Trigger trigger = TriggerBuilder
                .newTrigger()
                .withIdentity("cert-updater-trigger")
                .withSchedule(CronScheduleBuilder.cronSchedule("0/20 * * * * ?"))
                .build();

        try {
            Scheduler scheduler = new StdSchedulerFactory().getScheduler();
            scheduler.start();
            scheduler.scheduleJob(job, trigger);
        } catch (SchedulerException e) {
            e.printStackTrace();
        }
    }
}
